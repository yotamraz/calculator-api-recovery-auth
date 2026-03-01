package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// testConfig returns a Config suitable for testing.
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

// registerTestUser registers a user via the API and returns the HTTP response recorder.
func registerTestUser(router *gin.Engine, username, password string) *httptest.ResponseRecorder {
	body := `{"username":"` + username + `","password":"` + password + `"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

// loginTestUser logs in a user via form-encoded POST and returns the HTTP response recorder.
func loginTestUser(router *gin.Engine, username, password string) *httptest.ResponseRecorder {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	return w
}

// getTestToken registers a user and obtains a JWT token for use in authenticated requests.
func getTestToken(router *gin.Engine, username, password string) string {
	registerTestUser(router, username, password)
	w := loginTestUser(router, username, password)
	var tok Token
	json.Unmarshal(w.Body.Bytes(), &tok)
	return tok.AccessToken
}

// ============================================================================
// Health Check Tests (preserved from milestone 1)
// ============================================================================

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
// Registration Tests
// ============================================================================

func TestRegisterSuccess(t *testing.T) {
	router := setupTestRouter()
	w := registerTestUser(router, "testuser", "testpass123")

	if w.Code != http.StatusCreated {
		t.Errorf("POST /auth/register status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Username != "testuser" {
		t.Errorf("username = %q, want %q", resp.Username, "testuser")
	}
	if resp.ID == 0 {
		t.Error("expected non-zero user ID")
	}
	if resp.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	router := setupTestRouter()

	// Register first user
	registerTestUser(router, "dupuser", "pass1")

	// Try to register with same username
	w := registerTestUser(router, "dupuser", "pass2")

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /auth/register (duplicate) status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if resp["detail"] != "Username already taken" {
		t.Errorf("error detail = %q, want %q", resp["detail"], "Username already taken")
	}
}

func TestRegisterInvalidBody(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /auth/register (empty body) status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ============================================================================
// Login / Token Tests
// ============================================================================

func TestLoginSuccess(t *testing.T) {
	router := setupTestRouter()

	// Register user first
	registerTestUser(router, "loginuser", "loginpass")

	// Login
	w := loginTestUser(router, "loginuser", "loginpass")

	if w.Code != http.StatusOK {
		t.Errorf("POST /auth/token status = %d, want %d", w.Code, http.StatusOK)
	}

	var tok Token
	if err := json.Unmarshal(w.Body.Bytes(), &tok); err != nil {
		t.Fatalf("Failed to unmarshal token response: %v", err)
	}

	if tok.AccessToken == "" {
		t.Error("expected non-empty access_token")
	}
	if tok.TokenType != "bearer" {
		t.Errorf("token_type = %q, want %q", tok.TokenType, "bearer")
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	router := setupTestRouter()

	registerTestUser(router, "loginuser2", "correctpass")
	w := loginTestUser(router, "loginuser2", "wrongpass")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token (bad password) status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Incorrect username or password" {
		t.Errorf("error detail = %q, want %q", resp["detail"], "Incorrect username or password")
	}
}

func TestLoginNonExistentUser(t *testing.T) {
	router := setupTestRouter()

	w := loginTestUser(router, "nouser", "nopass")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token (no user) status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Incorrect username or password" {
		t.Errorf("error detail = %q, want %q", resp["detail"], "Incorrect username or password")
	}
}

func TestLoginEmptyFields(t *testing.T) {
	router := setupTestRouter()

	form := url.Values{}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token (empty fields) status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============================================================================
// JWT Token Validation Tests
// ============================================================================

func TestCreateAndParseToken(t *testing.T) {
	cfg := testConfig()
	token, err := CreateAccessToken("testuser", cfg)
	if err != nil {
		t.Fatalf("CreateAccessToken failed: %v", err)
	}

	username, err := ParseToken(token, cfg.JWTSecretKey)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if username != "testuser" {
		t.Errorf("ParseToken username = %q, want %q", username, "testuser")
	}
}

func TestParseExpiredToken(t *testing.T) {
	cfg := testConfig()

	// Create a token that expired 1 hour ago
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   "expireduser",
		ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecretKey))
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = ParseToken(tokenString, cfg.JWTSecretKey)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestParseMalformedToken(t *testing.T) {
	cfg := testConfig()

	_, err := ParseToken("not.a.valid.token", cfg.JWTSecretKey)
	if err == nil {
		t.Error("expected error for malformed token, got nil")
	}
}

func TestParseTokenWrongSecret(t *testing.T) {
	cfg := testConfig()
	token, err := CreateAccessToken("testuser", cfg)
	if err != nil {
		t.Fatalf("CreateAccessToken failed: %v", err)
	}

	_, err = ParseToken(token, "wrong-secret-key")
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

func TestParseTokenMissingSubject(t *testing.T) {
	cfg := testConfig()

	// Create a token without subject claim
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(30 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecretKey))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	_, err = ParseToken(tokenString, cfg.JWTSecretKey)
	if err == nil {
		t.Error("expected error for missing subject, got nil")
	}
}

// ============================================================================
// Auth Middleware Tests
// ============================================================================

func TestMiddlewareMissingToken(t *testing.T) {
	// Test the middleware directly by creating a temporary protected route.
	// Since no protected routes are registered yet (milestone 3), we build
	// a minimal router with the middleware applied.
	cfg := testConfig()
	db := InitDB(":memory:")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(db, cfg))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("protected route (no token) status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Could not validate credentials" {
		t.Errorf("error detail = %q, want %q", resp["detail"], "Could not validate credentials")
	}
}

func TestMiddlewareInvalidToken(t *testing.T) {
	cfg := testConfig()
	db := InitDB(":memory:")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(db, cfg))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("protected route (invalid token) status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Could not validate credentials" {
		t.Errorf("error detail = %q, want %q", resp["detail"], "Could not validate credentials")
	}
}

func TestMiddlewareExpiredToken(t *testing.T) {
	cfg := testConfig()
	db := InitDB(":memory:")
	db.AutoMigrate(&User{})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(db, cfg))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Create an expired token
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   "testuser",
		ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecretKey))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("protected route (expired token) status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Could not validate credentials" {
		t.Errorf("error detail = %q, want %q", resp["detail"], "Could not validate credentials")
	}
}

func TestMiddlewareValidToken(t *testing.T) {
	cfg := testConfig()
	db := InitDB(":memory:")
	db.AutoMigrate(&User{})

	// Create a user in the database
	hashed, _ := HashPassword("testpass")
	db.Create(&User{Username: "testuser", HashedPassword: hashed})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(db, cfg))
	r.GET("/protected", func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "no user in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"username": user.Username})
	})

	// Create a valid token
	tokenString, _ := CreateAccessToken("testuser", cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("protected route (valid token) status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["username"] != "testuser" {
		t.Errorf("username in response = %q, want %q", resp["username"], "testuser")
	}
}

func TestMiddlewareUserNotInDB(t *testing.T) {
	cfg := testConfig()
	db := InitDB(":memory:")
	db.AutoMigrate(&User{})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(db, cfg))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Create a valid token for a user that doesn't exist in DB
	tokenString, _ := CreateAccessToken("ghostuser", cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("protected route (user not in DB) status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Could not validate credentials" {
		t.Errorf("error detail = %q, want %q", resp["detail"], "Could not validate credentials")
	}
}

func TestMiddlewareBadAuthScheme(t *testing.T) {
	cfg := testConfig()
	db := InitDB(":memory:")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(db, cfg))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("protected route (Basic auth) status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============================================================================
// Password Hashing Tests
// ============================================================================

func TestHashAndVerifyPassword(t *testing.T) {
	password := "mysecretpassword"
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !VerifyPassword(password, hashed) {
		t.Error("VerifyPassword returned false for correct password")
	}

	if VerifyPassword("wrongpassword", hashed) {
		t.Error("VerifyPassword returned true for incorrect password")
	}
}

// ============================================================================
// Full Auth Flow Integration Test
// ============================================================================

func TestFullAuthFlow(t *testing.T) {
	router := setupTestRouter()

	// Step 1: Register
	regW := registerTestUser(router, "flowuser", "flowpass")
	if regW.Code != http.StatusCreated {
		t.Fatalf("Registration failed: status %d", regW.Code)
	}

	// Step 2: Login
	loginW := loginTestUser(router, "flowuser", "flowpass")
	if loginW.Code != http.StatusOK {
		t.Fatalf("Login failed: status %d", loginW.Code)
	}

	var tok Token
	json.Unmarshal(loginW.Body.Bytes(), &tok)
	if tok.AccessToken == "" {
		t.Fatal("Expected non-empty access token")
	}
	if tok.TokenType != "bearer" {
		t.Errorf("token_type = %q, want %q", tok.TokenType, "bearer")
	}

	// Step 3: Verify the token can be parsed back to the correct username
	cfg := testConfig()
	username, err := ParseToken(tok.AccessToken, cfg.JWTSecretKey)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if username != "flowuser" {
		t.Errorf("Token subject = %q, want %q", username, "flowuser")
	}
}
