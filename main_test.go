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

// SetupTestRouter creates a Gin router with an in-memory SQLite database for testing.
// This helper is designed to be reusable by future milestone tests.
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}

	if err := database.AutoMigrate(&User{}, &Calculation{}); err != nil {
		panic("failed to migrate test database: " + err.Error())
	}

	return SetupRouter(database)
}

// registerTestUser is a helper that registers a user and returns the response recorder.
func registerTestUser(router *gin.Engine, username, password string) *httptest.ResponseRecorder {
	body := `{"username":"` + username + `","password":"` + password + `"}`
	req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// getTestToken is a helper that registers a user and returns a valid JWT token.
func getTestToken(router *gin.Engine, username, password string) string {
	// Register user
	registerTestUser(router, username, password)

	// Get token
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var token Token
	json.Unmarshal(w.Body.Bytes(), &token)
	return token.AccessToken
}

// ---- Health Check Tests ----

func TestHealthCheck(t *testing.T) {
	router := SetupTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /health status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("health status = %q, want %q", resp.Status, "ok")
	}
	if resp.Version != "0.1.0" {
		t.Errorf("health version = %q, want %q", resp.Version, "0.1.0")
	}
}

// ---- Auth Register Tests ----

func TestRegisterSuccess(t *testing.T) {
	router := SetupTestRouter()

	w := registerTestUser(router, "testuser", "testpassword")

	if w.Code != http.StatusCreated {
		t.Errorf("POST /auth/register status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Username != "testuser" {
		t.Errorf("username = %q, want %q", resp.Username, "testuser")
	}
	if resp.ID == 0 {
		t.Error("expected non-zero user ID")
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	router := SetupTestRouter()

	// Register first user
	registerTestUser(router, "testuser", "testpassword")

	// Try to register same username
	w := registerTestUser(router, "testuser", "otherpassword")

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /auth/register duplicate status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}
	if resp.Detail != "Username already taken" {
		t.Errorf("error detail = %q, want %q", resp.Detail, "Username already taken")
	}
}

// ---- Auth Token Tests ----

func TestTokenSuccess(t *testing.T) {
	router := SetupTestRouter()

	// Register user first
	registerTestUser(router, "testuser", "testpassword")

	// Get token
	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("password", "testpassword")
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /auth/token status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp Token
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access_token")
	}
	if resp.TokenType != "bearer" {
		t.Errorf("token_type = %q, want %q", resp.TokenType, "bearer")
	}
}

func TestTokenInvalidCredentials(t *testing.T) {
	router := SetupTestRouter()

	// Register user
	registerTestUser(router, "testuser", "testpassword")

	// Try wrong password
	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("password", "wrongpassword")
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token invalid creds status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}
	if resp.Detail != "Incorrect username or password" {
		t.Errorf("error detail = %q, want %q", resp.Detail, "Incorrect username or password")
	}
}

func TestTokenNonexistentUser(t *testing.T) {
	router := SetupTestRouter()

	form := url.Values{}
	form.Set("username", "nonexistent")
	form.Set("password", "somepassword")
	req, _ := http.NewRequest("POST", "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token nonexistent user status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ---- Calculator Endpoint Tests ----

func TestCalculatorEndpointsRequireAuth(t *testing.T) {
	router := SetupTestRouter()

	endpoints := []string{"/add", "/subtract", "/multiply", "/divide"}
	body := `{"a": 1, "b": 2}`

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			req, _ := http.NewRequest("POST", ep, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("POST %s without auth status = %d, want %d", ep, w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestAddEndpoint(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	body := `{"a": 3, "b": 5}`
	req, _ := http.NewRequest("POST", "/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /add status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ResultResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Result != 8 {
		t.Errorf("add result = %v, want 8", resp.Result)
	}
}

func TestSubtractEndpoint(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	body := `{"a": 10, "b": 3}`
	req, _ := http.NewRequest("POST", "/subtract", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /subtract status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ResultResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Result != 7 {
		t.Errorf("subtract result = %v, want 7", resp.Result)
	}
}

func TestMultiplyEndpoint(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	body := `{"a": 4, "b": 5}`
	req, _ := http.NewRequest("POST", "/multiply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /multiply status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ResultResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Result != 20 {
		t.Errorf("multiply result = %v, want 20", resp.Result)
	}
}

func TestDivideEndpoint(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	body := `{"a": 10, "b": 4}`
	req, _ := http.NewRequest("POST", "/divide", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /divide status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ResultResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Result != 2.5 {
		t.Errorf("divide result = %v, want 2.5", resp.Result)
	}
}

func TestDivideByZero(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	body := `{"a": 10, "b": 0}`
	req, _ := http.NewRequest("POST", "/divide", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /divide by zero status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Detail != "Cannot divide by zero" {
		t.Errorf("error detail = %q, want %q", resp.Detail, "Cannot divide by zero")
	}
}

// ---- Calculations CRUD Tests ----

func TestCalculationsCRUDRequireAuth(t *testing.T) {
	router := SetupTestRouter()

	// POST /calculations
	req, _ := http.NewRequest("POST", "/calculations", strings.NewReader(`{"operation":"add","a":1,"b":2}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /calculations without auth status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// GET /calculations
	req, _ = http.NewRequest("GET", "/calculations", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("GET /calculations without auth status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// GET /calculations/1
	req, _ = http.NewRequest("GET", "/calculations/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("GET /calculations/1 without auth status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// DELETE /calculations/1
	req, _ = http.NewRequest("DELETE", "/calculations/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("DELETE /calculations/1 without auth status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestCreateCalculation(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	body := `{"operation":"add","a":3,"b":5}`
	req, _ := http.NewRequest("POST", "/calculations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("POST /calculations status = %d, want %d. Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp CalculationResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Operation != "add" {
		t.Errorf("operation = %q, want %q", resp.Operation, "add")
	}
	if resp.Result != 8 {
		t.Errorf("result = %v, want 8", resp.Result)
	}
	if resp.ID == 0 {
		t.Error("expected non-zero calculation ID")
	}
}

func TestCreateCalculationUnknownOp(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	body := `{"operation":"modulo","a":3,"b":5}`
	req, _ := http.NewRequest("POST", "/calculations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /calculations unknown op status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateCalculationDivByZero(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	body := `{"operation":"div","a":3,"b":0}`
	req, _ := http.NewRequest("POST", "/calculations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /calculations div/0 status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Detail != "Cannot divide by zero" {
		t.Errorf("error detail = %q, want %q", resp.Detail, "Cannot divide by zero")
	}
}

func TestListCalculations(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	// Create two calculations
	for _, body := range []string{
		`{"operation":"add","a":1,"b":2}`,
		`{"operation":"mul","a":3,"b":4}`,
	} {
		req, _ := http.NewRequest("POST", "/calculations", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// List calculations
	req, _ := http.NewRequest("GET", "/calculations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /calculations status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp []CalculationResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 calculations, got %d", len(resp))
	}

	// Should be ordered by created_at descending (most recent first)
	if len(resp) == 2 && resp[0].CreatedAt.Before(resp[1].CreatedAt) {
		t.Error("calculations should be ordered by created_at descending")
	}
}

func TestGetCalculation(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	// Create a calculation
	body := `{"operation":"sub","a":10,"b":3}`
	req, _ := http.NewRequest("POST", "/calculations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var created CalculationResponse
	json.Unmarshal(w.Body.Bytes(), &created)

	// Fetch by ID
	req, _ = http.NewRequest("GET", "/calculations/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /calculations/1 status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp CalculationResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Operation != "sub" || resp.Result != 7 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetCalculationNotFound(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	req, _ := http.NewRequest("GET", "/calculations/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /calculations/999 status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Detail != "Calculation not found" {
		t.Errorf("error detail = %q, want %q", resp.Detail, "Calculation not found")
	}
}

func TestDeleteCalculation(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	// Create a calculation
	body := `{"operation":"add","a":1,"b":2}`
	req, _ := http.NewRequest("POST", "/calculations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Delete it
	req, _ = http.NewRequest("DELETE", "/calculations/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("DELETE /calculations/1 status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify it's gone
	req, _ = http.NewRequest("GET", "/calculations/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /calculations/1 after delete status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestDeleteCalculationNotFound(t *testing.T) {
	router := SetupTestRouter()
	token := getTestToken(router, "testuser", "testpassword")

	req, _ := http.NewRequest("DELETE", "/calculations/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("DELETE /calculations/999 status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
