package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// testConfig returns a Config suitable for testing with known, predictable values.
func testConfig() Config {
	return Config{
		JWTSecretKey:         "test-secret-key",
		DatabaseURL:          ":memory:",
		AccessTokenExpireMin: 30,
	}
}

// setupTestRouter creates a Gin engine with an in-memory SQLite database
// for integration testing.
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := testConfig()
	db := InitDB(cfg.DatabaseURL)
	return SetupRouter(db, cfg)
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /health status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("GET /health status = %q, want %q", resp.Status, "ok")
	}
	if resp.Version != "0.1.0" {
		t.Errorf("GET /health version = %q, want %q", resp.Version, "0.1.0")
	}
}

func TestHealthCheckContentType(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("GET /health Content-Type = %q, want %q", contentType, "application/json; charset=utf-8")
	}
}

// ============================================================================
// Auth Endpoint Integration Tests
// ============================================================================

// registerUser is a test helper that sends a POST /auth/register request.
func registerUser(router *gin.Engine, username, password string) *httptest.ResponseRecorder {
	body := `{"username":"` + username + `","password":"` + password + `"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

// loginUser is a test helper that sends a POST /auth/token request with form data.
func loginUser(router *gin.Engine, username, password string) *httptest.ResponseRecorder {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	return w
}

func TestRegisterSuccess(t *testing.T) {
	router := setupTestRouter()

	w := registerUser(router, "alice", "secretpass")

	if w.Code != http.StatusCreated {
		t.Errorf("POST /auth/register status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Username != "alice" {
		t.Errorf("Username = %q, want %q", resp.Username, "alice")
	}
	if resp.ID == 0 {
		t.Error("ID should be non-zero after creation")
	}
	if resp.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestRegisterResponseFormat(t *testing.T) {
	router := setupTestRouter()

	w := registerUser(router, "bob", "password123")

	if w.Code != http.StatusCreated {
		t.Fatalf("POST /auth/register status = %d, want %d", w.Code, http.StatusCreated)
	}

	// Verify JSON keys match Python API exactly
	var raw map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &raw)

	expectedKeys := []string{"id", "username", "created_at"}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("Response missing key %q", key)
		}
	}

	// Ensure password/hashed_password is NOT in response
	if _, ok := raw["password"]; ok {
		t.Error("Response should NOT contain 'password' field")
	}
	if _, ok := raw["hashed_password"]; ok {
		t.Error("Response should NOT contain 'hashed_password' field")
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	router := setupTestRouter()

	// First registration should succeed
	w1 := registerUser(router, "alice", "pass1")
	if w1.Code != http.StatusCreated {
		t.Fatalf("First registration status = %d, want %d", w1.Code, http.StatusCreated)
	}

	// Second registration with same username should fail
	w2 := registerUser(router, "alice", "pass2")

	if w2.Code != http.StatusBadRequest {
		t.Errorf("Duplicate registration status = %d, want %d", w2.Code, http.StatusBadRequest)
	}

	var resp map[string]string
	json.Unmarshal(w2.Body.Bytes(), &resp)
	if resp["detail"] != "Username already taken" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Username already taken")
	}
}

func TestRegisterInvalidBody(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Invalid body status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRegisterMissingFields(t *testing.T) {
	router := setupTestRouter()

	// Missing password
	body := `{"username":"alice"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Missing password status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestLoginSuccess(t *testing.T) {
	router := setupTestRouter()

	// Register first
	registerUser(router, "alice", "secretpass")

	// Login
	w := loginUser(router, "alice", "secretpass")

	if w.Code != http.StatusOK {
		t.Errorf("POST /auth/token status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp Token
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("access_token should not be empty")
	}
	if resp.TokenType != "bearer" {
		t.Errorf("token_type = %q, want %q", resp.TokenType, "bearer")
	}
}

func TestLoginResponseFormat(t *testing.T) {
	router := setupTestRouter()

	registerUser(router, "alice", "secretpass")
	w := loginUser(router, "alice", "secretpass")

	if w.Code != http.StatusOK {
		t.Fatalf("POST /auth/token status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify JSON keys match Python API exactly
	var raw map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &raw)

	if _, ok := raw["access_token"]; !ok {
		t.Error("Response missing key 'access_token'")
	}
	if _, ok := raw["token_type"]; !ok {
		t.Error("Response missing key 'token_type'")
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	router := setupTestRouter()

	registerUser(router, "alice", "correctpass")
	w := loginUser(router, "alice", "wrongpass")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token with wrong password status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Incorrect username or password" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Incorrect username or password")
	}

	wwwAuth := w.Header().Get("WWW-Authenticate")
	if wwwAuth != "Bearer" {
		t.Errorf("WWW-Authenticate = %q, want %q", wwwAuth, "Bearer")
	}
}

func TestLoginNonexistentUser(t *testing.T) {
	router := setupTestRouter()

	w := loginUser(router, "nonexistent", "anypass")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token with non-existent user status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Incorrect username or password" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Incorrect username or password")
	}

	wwwAuth := w.Header().Get("WWW-Authenticate")
	if wwwAuth != "Bearer" {
		t.Errorf("WWW-Authenticate = %q, want %q", wwwAuth, "Bearer")
	}
}

func TestLoginFormEncoded(t *testing.T) {
	router := setupTestRouter()

	registerUser(router, "formuser", "formpass")

	// Ensure form-encoded data works (as the Python OAuth2PasswordRequestForm expects)
	form := url.Values{}
	form.Set("username", "formuser")
	form.Set("password", "formpass")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Form-encoded login status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp Token
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.AccessToken == "" {
		t.Error("access_token should not be empty for form-encoded login")
	}
}

func TestLoginTokenIsValidJWT(t *testing.T) {
	router := setupTestRouter()
	cfg := testConfig()

	registerUser(router, "alice", "secretpass")
	w := loginUser(router, "alice", "secretpass")

	var resp Token
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Verify the returned token is a valid JWT that can be parsed
	token, err := jwt.ParseWithClaims(resp.AccessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecretKey), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse returned JWT: %v", err)
	}
	if !token.Valid {
		t.Error("Returned JWT should be valid")
	}

	claims := token.Claims.(*jwt.RegisteredClaims)
	if claims.Subject != "alice" {
		t.Errorf("JWT subject = %q, want %q", claims.Subject, "alice")
	}
}

func TestFullAuthFlow(t *testing.T) {
	router := setupTestRouter()

	// 1. Register a user
	regW := registerUser(router, "fullflow", "mypassword")
	if regW.Code != http.StatusCreated {
		t.Fatalf("Registration status = %d, want %d", regW.Code, http.StatusCreated)
	}

	// 2. Login to get a token
	loginW := loginUser(router, "fullflow", "mypassword")
	if loginW.Code != http.StatusOK {
		t.Fatalf("Login status = %d, want %d", loginW.Code, http.StatusOK)
	}

	var tokenResp Token
	json.Unmarshal(loginW.Body.Bytes(), &tokenResp)

	// 3. Verify the token works for identity (subject matches username)
	cfg := testConfig()
	token, _ := jwt.ParseWithClaims(tokenResp.AccessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecretKey), nil
	})
	claims := token.Claims.(*jwt.RegisteredClaims)
	if claims.Subject != "fullflow" {
		t.Errorf("JWT subject = %q, want %q", claims.Subject, "fullflow")
	}
}

func TestMultipleUsersRegistration(t *testing.T) {
	router := setupTestRouter()

	// Register multiple distinct users
	w1 := registerUser(router, "user1", "pass1")
	w2 := registerUser(router, "user2", "pass2")
	w3 := registerUser(router, "user3", "pass3")

	if w1.Code != http.StatusCreated {
		t.Errorf("User1 registration status = %d, want %d", w1.Code, http.StatusCreated)
	}
	if w2.Code != http.StatusCreated {
		t.Errorf("User2 registration status = %d, want %d", w2.Code, http.StatusCreated)
	}
	if w3.Code != http.StatusCreated {
		t.Errorf("User3 registration status = %d, want %d", w3.Code, http.StatusCreated)
	}

	// Verify each can login independently
	l1 := loginUser(router, "user1", "pass1")
	l2 := loginUser(router, "user2", "pass2")
	l3 := loginUser(router, "user3", "pass3")

	if l1.Code != http.StatusOK {
		t.Errorf("User1 login status = %d, want %d", l1.Code, http.StatusOK)
	}
	if l2.Code != http.StatusOK {
		t.Errorf("User2 login status = %d, want %d", l2.Code, http.StatusOK)
	}
	if l3.Code != http.StatusOK {
		t.Errorf("User3 login status = %d, want %d", l3.Code, http.StatusOK)
	}
}
