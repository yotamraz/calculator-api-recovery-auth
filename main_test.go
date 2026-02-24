package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestRouter creates a Gin router with an in-memory SQLite database for testing.
func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&User{}, &Calculation{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	router := setupRouter(db)
	return router, db
}

// registerAndLogin is a test helper that registers a user and returns a valid JWT token.
func registerAndLogin(t *testing.T, router *gin.Engine) string {
	t.Helper()

	// Register user
	body := `{"username":"testuser","password":"testpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Register failed with status %d: %s", w.Code, w.Body.String())
	}

	// Login to get token
	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("password", "testpass123")
	req = httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Login failed with status %d: %s", w.Code, w.Body.String())
	}

	var tokenResp Token
	if err := json.Unmarshal(w.Body.Bytes(), &tokenResp); err != nil {
		t.Fatalf("Failed to parse token response: %v", err)
	}

	return tokenResp.AccessToken
}

// --- Health Check Tests ---

func TestHealthCheck(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", resp.Status)
	}
	if resp.Version != "0.1.0" {
		t.Errorf("Expected version '0.1.0', got '%s'", resp.Version)
	}
}

// --- Auth Register Tests ---

func TestRegisterNewUser(t *testing.T) {
	router, _ := setupTestRouter(t)

	body := `{"username":"newuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Username != "newuser" {
		t.Errorf("Expected username 'newuser', got '%s'", resp.Username)
	}
	if resp.ID == 0 {
		t.Error("Expected non-zero user ID")
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Register first user
	body := `{"username":"dupuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("First registration failed: %d", w.Code)
	}

	// Try to register same username
	req = httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for duplicate, got %d", w.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}
	if errResp["detail"] != "Username already taken" {
		t.Errorf("Expected 'Username already taken', got '%s'", errResp["detail"])
	}
}

// --- Auth Token Tests ---

func TestLoginValidCredentials(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Register user first
	body := `{"username":"loginuser","password":"mypassword"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Registration failed: %d", w.Code)
	}

	// Login with form-encoded data
	form := url.Values{}
	form.Set("username", "loginuser")
	form.Set("password", "mypassword")
	req = httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var tokenResp Token
	if err := json.Unmarshal(w.Body.Bytes(), &tokenResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if tokenResp.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if tokenResp.TokenType != "bearer" {
		t.Errorf("Expected token_type 'bearer', got '%s'", tokenResp.TokenType)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Register user
	body := `{"username":"authuser","password":"correctpass"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Login with wrong password
	form := url.Values{}
	form.Set("username", "authuser")
	form.Set("password", "wrongpass")
	req = httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestLoginNonexistentUser(t *testing.T) {
	router, _ := setupTestRouter(t)

	form := url.Values{}
	form.Set("username", "nouser")
	form.Set("password", "anypass")
	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// --- Auth Middleware Tests ---

func TestProtectedEndpointWithoutToken(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Try to access /add without authentication
	body := `{"a":1,"b":2}`
	req := httptest.NewRequest(http.MethodPost, "/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestProtectedEndpointWithInvalidToken(t *testing.T) {
	router, _ := setupTestRouter(t)

	body := `{"a":1,"b":2}`
	req := httptest.NewRequest(http.MethodPost, "/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// --- Calculator Endpoint Tests ---

func TestAddEndpoint(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	body := `{"a":10,"b":5}`
	req := httptest.NewRequest(http.MethodPost, "/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ResultResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Result != 15 {
		t.Errorf("Expected result 15, got %v", resp.Result)
	}
}

func TestSubtractEndpoint(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	body := `{"a":10,"b":3}`
	req := httptest.NewRequest(http.MethodPost, "/subtract", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp ResultResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Result != 7 {
		t.Errorf("Expected result 7, got %v", resp.Result)
	}
}

func TestMultiplyEndpoint(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	body := `{"a":4,"b":5}`
	req := httptest.NewRequest(http.MethodPost, "/multiply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp ResultResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Result != 20 {
		t.Errorf("Expected result 20, got %v", resp.Result)
	}
}

func TestDivideEndpoint(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	body := `{"a":10,"b":4}`
	req := httptest.NewRequest(http.MethodPost, "/divide", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp ResultResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Result != 2.5 {
		t.Errorf("Expected result 2.5, got %v", resp.Result)
	}
}

func TestDivideByZeroEndpoint(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	body := `{"a":10,"b":0}`
	req := httptest.NewRequest(http.MethodPost, "/divide", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var errResp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}
	if errResp["detail"] != "Cannot divide by zero" {
		t.Errorf("Expected 'Cannot divide by zero', got '%s'", errResp["detail"])
	}
}

func TestCalculatorWithZeroValues(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	// Test that a=0 and b=0 are valid inputs for add
	body := `{"a":0,"b":5}`
	req := httptest.NewRequest(http.MethodPost, "/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ResultResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Result != 5 {
		t.Errorf("Expected result 5, got %v", resp.Result)
	}
}

// --- Calculations CRUD Tests ---

func TestCreateCalculation(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	body := `{"operation":"add","a":10,"b":5}`
	req := httptest.NewRequest(http.MethodPost, "/calculations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp CalculationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Operation != "add" {
		t.Errorf("Expected operation 'add', got '%s'", resp.Operation)
	}
	if resp.Result != 15 {
		t.Errorf("Expected result 15, got %v", resp.Result)
	}
}

func TestCreateCalculationUnknownOperation(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	body := `{"operation":"modulo","a":10,"b":3}`
	req := httptest.NewRequest(http.MethodPost, "/calculations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestListCalculations(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	// Create two calculations
	for _, op := range []string{"add", "sub"} {
		body := `{"operation":"` + op + `","a":10,"b":5}`
		req := httptest.NewRequest(http.MethodPost, "/calculations", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("Failed to create calculation: %d", w.Code)
		}
	}

	// List calculations
	req := httptest.NewRequest(http.MethodGet, "/calculations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp []CalculationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("Expected 2 calculations, got %d", len(resp))
	}

	// Verify ordering (most recent first)
	if len(resp) == 2 && resp[0].CreatedAt.Before(resp[1].CreatedAt) {
		t.Error("Expected calculations ordered by created_at descending")
	}
}

func TestGetCalculation(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	// Create a calculation
	body := `{"operation":"mul","a":3,"b":4}`
	req := httptest.NewRequest(http.MethodPost, "/calculations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var created CalculationResponse
	json.Unmarshal(w.Body.Bytes(), &created)

	// Get it by ID
	req = httptest.NewRequest(http.MethodGet, "/calculations/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp CalculationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Result != 12 {
		t.Errorf("Expected result 12, got %v", resp.Result)
	}
}

func TestGetCalculationNotFound(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	req := httptest.NewRequest(http.MethodGet, "/calculations/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestDeleteCalculation(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	// Create a calculation
	body := `{"operation":"add","a":1,"b":1}`
	req := httptest.NewRequest(http.MethodPost, "/calculations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create calculation: %d", w.Code)
	}

	// Delete it
	req = httptest.NewRequest(http.MethodDelete, "/calculations/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/calculations/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 after delete, got %d", w.Code)
	}
}

func TestDeleteCalculationNotFound(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := registerAndLogin(t, router)

	req := httptest.NewRequest(http.MethodDelete, "/calculations/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// --- Calculator Endpoint Auth Required Tests ---

func TestAllCalculatorEndpointsRequireAuth(t *testing.T) {
	router, _ := setupTestRouter(t)

	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/add", `{"a":1,"b":2}`},
		{http.MethodPost, "/subtract", `{"a":1,"b":2}`},
		{http.MethodPost, "/multiply", `{"a":1,"b":2}`},
		{http.MethodPost, "/divide", `{"a":1,"b":2}`},
		{http.MethodPost, "/calculations", `{"operation":"add","a":1,"b":2}`},
		{http.MethodGet, "/calculations", ""},
		{http.MethodGet, "/calculations/1", ""},
		{http.MethodDelete, "/calculations/1", ""},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			var req *http.Request
			if ep.body != "" {
				req = httptest.NewRequest(ep.method, ep.path, strings.NewReader(ep.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(ep.method, ep.path, nil)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected 401 for unauthenticated %s %s, got %d", ep.method, ep.path, w.Code)
			}
		})
	}
}
