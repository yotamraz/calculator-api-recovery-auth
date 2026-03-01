package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
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

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Invalid body status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
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

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Missing password status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
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

// ============================================================================
// Calculator Endpoint Integration Tests
// ============================================================================

// getAuthToken is a test helper that registers a user and returns a valid JWT token.
func getAuthToken(t *testing.T, router *gin.Engine) string {
	t.Helper()
	registerUser(router, "testuser", "testpass")
	w := loginUser(router, "testuser", "testpass")
	if w.Code != http.StatusOK {
		t.Fatalf("Login failed: status = %d", w.Code)
	}
	var resp Token
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.AccessToken
}

// authRequest is a test helper that sends an authenticated JSON request.
func authRequest(router *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	return w
}

func TestAddEndpoint(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "POST", "/add", `{"a": 5, "b": 3}`, token)

	if w.Code != http.StatusOK {
		t.Errorf("POST /add status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ResultResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Result != 8.0 {
		t.Errorf("POST /add result = %f, want %f", resp.Result, 8.0)
	}
}

func TestSubtractEndpoint(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "POST", "/subtract", `{"a": 10, "b": 3}`, token)

	if w.Code != http.StatusOK {
		t.Errorf("POST /subtract status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ResultResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Result != 7.0 {
		t.Errorf("POST /subtract result = %f, want %f", resp.Result, 7.0)
	}
}

func TestMultiplyEndpoint(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "POST", "/multiply", `{"a": 7, "b": 6}`, token)

	if w.Code != http.StatusOK {
		t.Errorf("POST /multiply status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ResultResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Result != 42.0 {
		t.Errorf("POST /multiply result = %f, want %f", resp.Result, 42.0)
	}
}

func TestDivideEndpoint(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "POST", "/divide", `{"a": 10, "b": 2}`, token)

	if w.Code != http.StatusOK {
		t.Errorf("POST /divide status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ResultResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Result != 5.0 {
		t.Errorf("POST /divide result = %f, want %f", resp.Result, 5.0)
	}
}

func TestDivideByZeroEndpoint(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "POST", "/divide", `{"a": 10, "b": 0}`, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /divide by zero status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Cannot divide by zero" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Cannot divide by zero")
	}
}

func TestCalculatorEndpointUnauthenticated(t *testing.T) {
	router := setupTestRouter()

	endpoints := []string{"/add", "/subtract", "/multiply", "/divide"}
	for _, ep := range endpoints {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", ep, strings.NewReader(`{"a": 1, "b": 1}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("POST %s unauthenticated status = %d, want %d", ep, w.Code, http.StatusUnauthorized)
		}

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["detail"] != "Could not validate credentials" {
			t.Errorf("POST %s unauthenticated detail = %q, want %q", ep, resp["detail"], "Could not validate credentials")
		}
	}
}

func TestCalculatorEndpointInvalidToken(t *testing.T) {
	router := setupTestRouter()

	w := authRequest(router, "POST", "/add", `{"a": 1, "b": 1}`, "invalid-token")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /add with invalid token status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Could not validate credentials" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Could not validate credentials")
	}
}

func TestAddNegativeNumbers(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "POST", "/add", `{"a": -5, "b": -3}`, token)

	if w.Code != http.StatusOK {
		t.Errorf("POST /add status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ResultResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Result != -8.0 {
		t.Errorf("POST /add result = %f, want %f", resp.Result, -8.0)
	}
}

func TestDivideDecimalResult(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "POST", "/divide", `{"a": 7, "b": 2}`, token)

	if w.Code != http.StatusOK {
		t.Errorf("POST /divide status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ResultResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Result != 3.5 {
		t.Errorf("POST /divide result = %f, want %f", resp.Result, 3.5)
	}
}

// ============================================================================
// Calculation CRUD Endpoint Integration Tests
// ============================================================================

func TestCreateCalculation(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "POST", "/calculations", `{"operation": "mul", "a": 7, "b": 6}`, token)

	if w.Code != http.StatusCreated {
		t.Errorf("POST /calculations status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp CalculationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Operation != "mul" {
		t.Errorf("Operation = %q, want %q", resp.Operation, "mul")
	}
	if resp.A != 7.0 {
		t.Errorf("A = %f, want %f", resp.A, 7.0)
	}
	if resp.B != 6.0 {
		t.Errorf("B = %f, want %f", resp.B, 6.0)
	}
	if resp.Result != 42.0 {
		t.Errorf("Result = %f, want %f", resp.Result, 42.0)
	}
	if resp.ID == 0 {
		t.Error("ID should be non-zero after creation")
	}
	if resp.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestCreateCalculationResponseFormat(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "POST", "/calculations", `{"operation": "add", "a": 1, "b": 2}`, token)

	if w.Code != http.StatusCreated {
		t.Fatalf("POST /calculations status = %d, want %d", w.Code, http.StatusCreated)
	}

	// Verify JSON keys match Python API exactly
	var raw map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &raw)

	expectedKeys := []string{"id", "operation", "a", "b", "result", "created_at"}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("Response missing key %q", key)
		}
	}
}

func TestCreateCalculationAllOperations(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	tests := []struct {
		op     string
		a, b   float64
		result float64
	}{
		{"add", 5, 3, 8},
		{"sub", 10, 4, 6},
		{"mul", 7, 6, 42},
		{"div", 10, 2, 5},
	}

	for _, tt := range tests {
		body := `{"operation": "` + tt.op + `", "a": ` + formatFloat(tt.a) + `, "b": ` + formatFloat(tt.b) + `}`
		w := authRequest(router, "POST", "/calculations", body, token)

		if w.Code != http.StatusCreated {
			t.Errorf("POST /calculations op=%s status = %d, want %d", tt.op, w.Code, http.StatusCreated)
			continue
		}

		var resp CalculationResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Result != tt.result {
			t.Errorf("POST /calculations op=%s result = %f, want %f", tt.op, resp.Result, tt.result)
		}
	}
}

func TestCreateCalculationUnknownOperation(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "POST", "/calculations", `{"operation": "mod", "a": 7, "b": 3}`, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /calculations unknown op status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	expected := "Unknown operation: mod. Use: ['add', 'sub', 'mul', 'div']"
	if resp["detail"] != expected {
		t.Errorf("Error detail = %q, want %q", resp["detail"], expected)
	}
}

func TestCreateCalculationDivideByZero(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "POST", "/calculations", `{"operation": "div", "a": 10, "b": 0}`, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /calculations div by zero status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Cannot divide by zero" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Cannot divide by zero")
	}
}

func TestCreateCalculationUnauthenticated(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/calculations", strings.NewReader(`{"operation": "add", "a": 1, "b": 2}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /calculations unauthenticated status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestListCalculations(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	// Create some calculations
	authRequest(router, "POST", "/calculations", `{"operation": "add", "a": 1, "b": 2}`, token)
	authRequest(router, "POST", "/calculations", `{"operation": "mul", "a": 3, "b": 4}`, token)

	// List all calculations
	w := authRequest(router, "GET", "/calculations", "", token)

	if w.Code != http.StatusOK {
		t.Errorf("GET /calculations status = %d, want %d", w.Code, http.StatusOK)
	}

	var calculations []CalculationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &calculations); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(calculations) != 2 {
		t.Errorf("GET /calculations count = %d, want %d", len(calculations), 2)
	}
}

func TestListCalculationsEmpty(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "GET", "/calculations", "", token)

	if w.Code != http.StatusOK {
		t.Errorf("GET /calculations empty status = %d, want %d", w.Code, http.StatusOK)
	}

	var calculations []CalculationResponse
	json.Unmarshal(w.Body.Bytes(), &calculations)
	if len(calculations) != 0 {
		t.Errorf("GET /calculations empty count = %d, want %d", len(calculations), 0)
	}
}

func TestListCalculationsOrder(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	// Create calculations in order
	authRequest(router, "POST", "/calculations", `{"operation": "add", "a": 1, "b": 1}`, token)
	authRequest(router, "POST", "/calculations", `{"operation": "sub", "a": 5, "b": 3}`, token)
	authRequest(router, "POST", "/calculations", `{"operation": "mul", "a": 2, "b": 3}`, token)

	// List and verify order (newest first)
	w := authRequest(router, "GET", "/calculations", "", token)

	var calculations []CalculationResponse
	json.Unmarshal(w.Body.Bytes(), &calculations)

	if len(calculations) != 3 {
		t.Fatalf("Expected 3 calculations, got %d", len(calculations))
	}

	// The most recent should be first (mul), then sub, then add
	if calculations[0].Operation != "mul" {
		t.Errorf("First calculation op = %q, want %q (newest first)", calculations[0].Operation, "mul")
	}
	if calculations[2].Operation != "add" {
		t.Errorf("Last calculation op = %q, want %q (oldest last)", calculations[2].Operation, "add")
	}
}

func TestGetCalculation(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	// Create a calculation
	createW := authRequest(router, "POST", "/calculations", `{"operation": "mul", "a": 7, "b": 6}`, token)
	var created CalculationResponse
	json.Unmarshal(createW.Body.Bytes(), &created)

	// Get the calculation by ID
	w := authRequest(router, "GET", "/calculations/"+itoa(created.ID), "", token)

	if w.Code != http.StatusOK {
		t.Errorf("GET /calculations/%d status = %d, want %d", created.ID, w.Code, http.StatusOK)
	}

	var resp CalculationResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.ID != created.ID {
		t.Errorf("ID = %d, want %d", resp.ID, created.ID)
	}
	if resp.Operation != "mul" {
		t.Errorf("Operation = %q, want %q", resp.Operation, "mul")
	}
	if resp.Result != 42.0 {
		t.Errorf("Result = %f, want %f", resp.Result, 42.0)
	}
}

func TestGetCalculationNotFound(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "GET", "/calculations/99999", "", token)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /calculations/99999 status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Calculation not found" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Calculation not found")
	}
}

func TestDeleteCalculation(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	// Create a calculation
	createW := authRequest(router, "POST", "/calculations", `{"operation": "add", "a": 5, "b": 3}`, token)
	var created CalculationResponse
	json.Unmarshal(createW.Body.Bytes(), &created)

	// Delete the calculation
	w := authRequest(router, "DELETE", "/calculations/"+itoa(created.ID), "", token)

	if w.Code != http.StatusNoContent {
		t.Errorf("DELETE /calculations/%d status = %d, want %d", created.ID, w.Code, http.StatusNoContent)
	}

	// Verify body is empty
	if w.Body.Len() != 0 {
		t.Errorf("DELETE response body should be empty, got %q", w.Body.String())
	}

	// Verify the calculation is gone
	getW := authRequest(router, "GET", "/calculations/"+itoa(created.ID), "", token)
	if getW.Code != http.StatusNotFound {
		t.Errorf("GET deleted calculation status = %d, want %d", getW.Code, http.StatusNotFound)
	}
}

func TestDeleteCalculationNotFound(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	w := authRequest(router, "DELETE", "/calculations/99999", "", token)

	if w.Code != http.StatusNotFound {
		t.Errorf("DELETE /calculations/99999 status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Calculation not found" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Calculation not found")
	}
}

func TestCRUDEndpointsUnauthenticated(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		method string
		path   string
	}{
		{"POST", "/calculations"},
		{"GET", "/calculations"},
		{"GET", "/calculations/1"},
		{"DELETE", "/calculations/1"},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(tt.method, tt.path, strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s unauthenticated status = %d, want %d", tt.method, tt.path, w.Code, http.StatusUnauthorized)
		}
	}
}

func TestFullCalculationCRUDLifecycle(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken(t, router)

	// 1. Create a calculation
	createW := authRequest(router, "POST", "/calculations", `{"operation": "add", "a": 10, "b": 20}`, token)
	if createW.Code != http.StatusCreated {
		t.Fatalf("Create status = %d, want %d", createW.Code, http.StatusCreated)
	}
	var created CalculationResponse
	json.Unmarshal(createW.Body.Bytes(), &created)

	// 2. Verify it appears in the list
	listW := authRequest(router, "GET", "/calculations", "", token)
	if listW.Code != http.StatusOK {
		t.Fatalf("List status = %d, want %d", listW.Code, http.StatusOK)
	}
	var list []CalculationResponse
	json.Unmarshal(listW.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("List count = %d, want %d", len(list), 1)
	}
	if list[0].ID != created.ID {
		t.Errorf("List[0].ID = %d, want %d", list[0].ID, created.ID)
	}

	// 3. Get it by ID
	getW := authRequest(router, "GET", "/calculations/"+itoa(created.ID), "", token)
	if getW.Code != http.StatusOK {
		t.Fatalf("Get status = %d, want %d", getW.Code, http.StatusOK)
	}
	var fetched CalculationResponse
	json.Unmarshal(getW.Body.Bytes(), &fetched)
	if fetched.Result != 30.0 {
		t.Errorf("Result = %f, want %f", fetched.Result, 30.0)
	}

	// 4. Delete it
	delW := authRequest(router, "DELETE", "/calculations/"+itoa(created.ID), "", token)
	if delW.Code != http.StatusNoContent {
		t.Fatalf("Delete status = %d, want %d", delW.Code, http.StatusNoContent)
	}

	// 5. Verify it's gone from the list
	listW2 := authRequest(router, "GET", "/calculations", "", token)
	var list2 []CalculationResponse
	json.Unmarshal(listW2.Body.Bytes(), &list2)
	if len(list2) != 0 {
		t.Errorf("List after delete count = %d, want %d", len(list2), 0)
	}

	// 6. Verify get returns 404
	getW2 := authRequest(router, "GET", "/calculations/"+itoa(created.ID), "", token)
	if getW2.Code != http.StatusNotFound {
		t.Errorf("Get deleted status = %d, want %d", getW2.Code, http.StatusNotFound)
	}
}

func TestFullAuthAndCalculatorFlow(t *testing.T) {
	router := setupTestRouter()

	// 1. Register
	regW := registerUser(router, "calcuser", "calcpass")
	if regW.Code != http.StatusCreated {
		t.Fatalf("Registration status = %d, want %d", regW.Code, http.StatusCreated)
	}

	// 2. Login
	loginW := loginUser(router, "calcuser", "calcpass")
	if loginW.Code != http.StatusOK {
		t.Fatalf("Login status = %d, want %d", loginW.Code, http.StatusOK)
	}
	var tokenResp Token
	json.Unmarshal(loginW.Body.Bytes(), &tokenResp)
	token := tokenResp.AccessToken

	// 3. Use calculator endpoints
	addW := authRequest(router, "POST", "/add", `{"a": 100, "b": 200}`, token)
	if addW.Code != http.StatusOK {
		t.Errorf("Add status = %d, want %d", addW.Code, http.StatusOK)
	}
	var addResp ResultResponse
	json.Unmarshal(addW.Body.Bytes(), &addResp)
	if addResp.Result != 300.0 {
		t.Errorf("Add result = %f, want %f", addResp.Result, 300.0)
	}

	// 4. Create a persisted calculation
	calcW := authRequest(router, "POST", "/calculations", `{"operation": "sub", "a": 50, "b": 25}`, token)
	if calcW.Code != http.StatusCreated {
		t.Errorf("Create calculation status = %d, want %d", calcW.Code, http.StatusCreated)
	}
	var calcResp CalculationResponse
	json.Unmarshal(calcW.Body.Bytes(), &calcResp)
	if calcResp.Result != 25.0 {
		t.Errorf("Calculation result = %f, want %f", calcResp.Result, 25.0)
	}

	// 5. List calculations
	listW := authRequest(router, "GET", "/calculations", "", token)
	if listW.Code != http.StatusOK {
		t.Errorf("List status = %d, want %d", listW.Code, http.StatusOK)
	}
	var list []CalculationResponse
	json.Unmarshal(listW.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Errorf("List count = %d, want %d", len(list), 1)
	}
}

// itoa converts an integer to a string (helper for building URL paths).
func itoa(n int) string {
	return strconv.Itoa(n)
}

// formatFloat converts a float64 to a string for JSON body construction.
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
