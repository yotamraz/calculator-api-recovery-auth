package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

// setupTestRouterWithDB creates a Gin engine and also returns the DB for
// test data setup.
func setupTestRouterWithDB() (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	cfg := testConfig()
	db := InitDB(cfg.DatabaseURL)
	return SetupRouter(db, cfg), db
}

// registerTestUser is a helper that registers a user via the API and returns
// the HTTP response recorder.
func registerTestUser(router *gin.Engine, username, password string) *httptest.ResponseRecorder {
	body := `{"username":"` + username + `","password":"` + password + `"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

// getTestToken is a helper that obtains a JWT token via the API.
func getTestToken(router *gin.Engine, username, password string) string {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	var token Token
	json.Unmarshal(w.Body.Bytes(), &token)
	return token.AccessToken
}

// ============================================================================
// Health Check Tests
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

	w := registerTestUser(router, "newuser", "password123")

	if w.Code != http.StatusCreated {
		t.Errorf("POST /auth/register status = %d, want %d. Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Username != "newuser" {
		t.Errorf("Response username = %q, want %q", resp.Username, "newuser")
	}
	if resp.ID == 0 {
		t.Error("Response ID should not be 0")
	}
	if resp.CreatedAt.IsZero() {
		t.Error("Response CreatedAt should not be zero")
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	router := setupTestRouter()

	// Register first user
	registerTestUser(router, "duplicate", "password123")

	// Try to register again with the same username
	w := registerTestUser(router, "duplicate", "password456")

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /auth/register duplicate status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Username already taken" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Username already taken")
	}
}

func TestRegisterMissingUsername(t *testing.T) {
	router := setupTestRouter()

	body := `{"password":"password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("POST /auth/register missing username status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
}

func TestRegisterMissingPassword(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"testuser"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("POST /auth/register missing password status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
}

// ============================================================================
// Token Tests
// ============================================================================

func TestTokenSuccess(t *testing.T) {
	router := setupTestRouter()

	// Register a user first
	registerTestUser(router, "tokenuser", "password123")

	// Request a token
	form := url.Values{}
	form.Set("username", "tokenuser")
	form.Set("password", "password123")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /auth/token status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp Token
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("Access token should not be empty")
	}
	if resp.TokenType != "bearer" {
		t.Errorf("Token type = %q, want %q", resp.TokenType, "bearer")
	}

	// Verify the token is valid and contains the correct username
	username, err := ParseToken(resp.AccessToken, "test-secret-key")
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if username != "tokenuser" {
		t.Errorf("Token username = %q, want %q", username, "tokenuser")
	}
}

func TestTokenInvalidPassword(t *testing.T) {
	router := setupTestRouter()

	// Register a user
	registerTestUser(router, "tokenuser2", "password123")

	// Try with wrong password
	form := url.Values{}
	form.Set("username", "tokenuser2")
	form.Set("password", "wrongpassword")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token wrong password status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Incorrect username or password" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Incorrect username or password")
	}

	// Check WWW-Authenticate header
	wwwAuth := w.Header().Get("WWW-Authenticate")
	if wwwAuth != "Bearer" {
		t.Errorf("WWW-Authenticate header = %q, want %q", wwwAuth, "Bearer")
	}
}

func TestTokenNonExistentUser(t *testing.T) {
	router := setupTestRouter()

	form := url.Values{}
	form.Set("username", "nonexistent")
	form.Set("password", "password123")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token non-existent user status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Incorrect username or password" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Incorrect username or password")
	}
}

func TestTokenEmptyFields(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token empty fields status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============================================================================
// Auth Middleware Tests
// ============================================================================

func TestProtectedRouteWithoutToken(t *testing.T) {
	// Test middleware directly with a custom test route.
	cfg := testConfig()
	db := InitDB(":memory:")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg.JWTSecretKey))
	protected.GET("/test-protected", func(c *gin.Context) {
		user := GetCurrentUser(c)
		c.JSON(http.StatusOK, gin.H{"username": user.Username})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Protected route without token status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Could not validate credentials" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Could not validate credentials")
	}

	wwwAuth := w.Header().Get("WWW-Authenticate")
	if wwwAuth != "Bearer" {
		t.Errorf("WWW-Authenticate header = %q, want %q", wwwAuth, "Bearer")
	}
}

func TestProtectedRouteWithInvalidToken(t *testing.T) {
	cfg := testConfig()
	db := InitDB(":memory:")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg.JWTSecretKey))
	protected.GET("/test-protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Protected route with invalid token status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Could not validate credentials" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Could not validate credentials")
	}
}

func TestProtectedRouteWithValidToken(t *testing.T) {
	cfg := testConfig()
	db := InitDB(":memory:")

	// Create a user in the database
	hash, _ := HashPassword("password123")
	db.Create(&User{Username: "authuser", HashedPassword: hash})

	// Create a valid token
	tokenString, _ := CreateAccessToken("authuser", cfg.JWTSecretKey, cfg.AccessTokenExpireMin)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg.JWTSecretKey))
	protected.GET("/test-protected", func(c *gin.Context) {
		user := GetCurrentUser(c)
		c.JSON(http.StatusOK, gin.H{"username": user.Username})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Protected route with valid token status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["username"] != "authuser" {
		t.Errorf("Response username = %q, want %q", resp["username"], "authuser")
	}
}

func TestProtectedRouteWithExpiredToken(t *testing.T) {
	cfg := testConfig()
	db := InitDB(":memory:")

	hash, _ := HashPassword("password123")
	db.Create(&User{Username: "authuser", HashedPassword: hash})

	// Create an expired token (negative expiry)
	tokenString, _ := CreateAccessToken("authuser", cfg.JWTSecretKey, -1)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg.JWTSecretKey))
	protected.GET("/test-protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Protected route with expired token status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestProtectedRouteWithMalformedAuthHeader(t *testing.T) {
	cfg := testConfig()
	db := InitDB(":memory:")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg.JWTSecretKey))
	protected.GET("/test-protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Test with "Basic" instead of "Bearer"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-protected", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Protected route with Basic auth status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestProtectedRouteTokenForNonExistentUser(t *testing.T) {
	cfg := testConfig()
	db := InitDB(":memory:")

	// Create a token for a user that doesn't exist in DB
	tokenString, _ := CreateAccessToken("ghost", cfg.JWTSecretKey, cfg.AccessTokenExpireMin)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg.JWTSecretKey))
	protected.GET("/test-protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Protected route with token for non-existent user status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Could not validate credentials" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Could not validate credentials")
	}
}

// ============================================================================
// Full Auth Flow Integration Test
// ============================================================================

func TestFullAuthFlow(t *testing.T) {
	router := setupTestRouter()

	// Step 1: Register a user
	w := registerTestUser(router, "flowuser", "securepass")
	if w.Code != http.StatusCreated {
		t.Fatalf("Registration failed: status = %d, body = %s", w.Code, w.Body.String())
	}

	// Step 2: Get a token
	token := getTestToken(router, "flowuser", "securepass")
	if token == "" {
		t.Fatal("Failed to get token")
	}

	// Step 3: Verify token is valid by parsing it
	username, err := ParseToken(token, "test-secret-key")
	if err != nil {
		t.Fatalf("Token is invalid: %v", err)
	}
	if username != "flowuser" {
		t.Errorf("Token username = %q, want %q", username, "flowuser")
	}
}
