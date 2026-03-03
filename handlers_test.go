package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestDeps creates a Deps instance backed by an in-memory SQLite database,
// suitable for integration tests. Each call creates a fresh, isolated database.
func newTestDeps(t *testing.T) *Deps {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	if err := db.AutoMigrate(&User{}, &Calculation{}); err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}
	return &Deps{
		DB: db,
		Config: Config{
			JWTSecretKey: "test-secret",
			ServerAddr:   ":0",
			DatabaseURL:  "file::memory:",
		},
	}
}

// newTestRouter sets up a Chi router with all routes, mirroring main.go wiring.
func newTestRouter(deps *Deps) *chi.Mux {
	r := chi.NewRouter()

	// Public routes.
	r.Get("/health", deps.HealthHandler)
	r.Post("/auth/register", deps.RegisterHandler)
	r.Post("/auth/token", deps.TokenHandler)

	// Protected routes.
	r.Group(func(pr chi.Router) {
		pr.Use(deps.AuthMiddleware)
		// Dummy protected endpoint for testing auth middleware via HTTP.
		pr.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			u := UserFromContext(r.Context())
			writeJSON(w, http.StatusOK, map[string]string{"username": u.Username})
		})
	})

	return r
}

// ---------------------------------------------------------------------------
// Health endpoint tests (from Milestone 1)
// ---------------------------------------------------------------------------

func TestHealthEndpoint(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify status code.
	if w.Code != http.StatusOK {
		t.Errorf("GET /health status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify Content-Type header.
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("GET /health Content-Type = %q, want %q", ct, "application/json")
	}

	// Verify response body.
	var resp HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("GET /health failed to decode body: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("GET /health status field = %q, want %q", resp.Status, "ok")
	}
	if resp.Version != "0.1.0" {
		t.Errorf("GET /health version field = %q, want %q", resp.Version, "0.1.0")
	}
}

func TestHealthEndpointMethod(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)

	// POST to /health should return 405 Method Not Allowed.
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /health status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestNotFoundRoute(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /nonexistent status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// ---------------------------------------------------------------------------
// Registration endpoint tests
// ---------------------------------------------------------------------------

func TestRegisterEndpoint(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)

	body := `{"username": "newuser", "password": "securepass123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("POST /auth/register status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp UserResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}
	if resp.Username != "newuser" {
		t.Errorf("register response username = %q, want %q", resp.Username, "newuser")
	}
	if resp.ID == 0 {
		t.Error("register response ID should not be 0")
	}
	if resp.CreatedAt.IsZero() {
		t.Error("register response created_at should not be zero")
	}
}

func TestRegisterEndpointResponseContentType(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)

	body := `{"username": "ctuser", "password": "pass"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)

	body := `{"username": "dupuser", "password": "pass123"}`

	// First registration should succeed.
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first register status = %d, want %d", w.Code, http.StatusCreated)
	}

	// Second registration with same username should fail.
	req = httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("duplicate register status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Detail != "Username already taken" {
		t.Errorf("error detail = %q, want %q", errResp.Detail, "Username already taken")
	}
}

func TestRegisterInvalidBody(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("invalid body status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
}

// ---------------------------------------------------------------------------
// Token endpoint tests
// ---------------------------------------------------------------------------

// registerTestUser is a helper that registers a user and returns the deps.
func registerTestUser(t *testing.T, deps *Deps, username, password string) {
	t.Helper()
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	user := User{Username: username, HashedPassword: hash}
	if err := deps.DB.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
}

func TestTokenEndpoint(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)
	registerTestUser(t, deps, "tokenuser", "mypassword")

	// POST form-encoded data.
	form := url.Values{}
	form.Set("username", "tokenuser")
	form.Set("password", "mypassword")

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /auth/token status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp Token
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("token response access_token should not be empty")
	}
	if resp.TokenType != "bearer" {
		t.Errorf("token response token_type = %q, want %q", resp.TokenType, "bearer")
	}

	// Verify the token is actually valid.
	username, err := ParseToken(resp.AccessToken, deps.Config.JWTSecretKey)
	if err != nil {
		t.Fatalf("returned token is invalid: %v", err)
	}
	if username != "tokenuser" {
		t.Errorf("token sub = %q, want %q", username, "tokenuser")
	}
}

func TestTokenEndpointInvalidCredentials(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)
	registerTestUser(t, deps, "tokenuser2", "correctpass")

	form := url.Values{}
	form.Set("username", "tokenuser2")
	form.Set("password", "wrongpass")

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// Check WWW-Authenticate header.
	if wwwAuth := w.Header().Get("WWW-Authenticate"); wwwAuth != "Bearer" {
		t.Errorf("WWW-Authenticate = %q, want %q", wwwAuth, "Bearer")
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Detail != "Incorrect username or password" {
		t.Errorf("error detail = %q, want %q", errResp.Detail, "Incorrect username or password")
	}
}

func TestTokenEndpointNonexistentUser(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)

	form := url.Values{}
	form.Set("username", "ghost")
	form.Set("password", "password")

	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/token status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ---------------------------------------------------------------------------
// Protected endpoint tests (auth middleware integration via HTTP)
// ---------------------------------------------------------------------------

func TestProtectedEndpointWithoutToken(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GET /protected without token status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// Check error message.
	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Detail != "Could not validate credentials" {
		t.Errorf("error detail = %q, want %q", errResp.Detail, "Could not validate credentials")
	}

	// Check WWW-Authenticate header.
	if wwwAuth := w.Header().Get("WWW-Authenticate"); wwwAuth != "Bearer" {
		t.Errorf("WWW-Authenticate = %q, want %q", wwwAuth, "Bearer")
	}
}

func TestProtectedEndpointWithInvalidToken(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GET /protected with invalid token status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestProtectedEndpointWithValidToken(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)
	registerTestUser(t, deps, "protecteduser", "pass123")

	token, err := CreateAccessToken("protecteduser", deps.Config.JWTSecretKey, 30)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /protected with valid token status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["username"] != "protecteduser" {
		t.Errorf("protected response username = %q, want %q", resp["username"], "protecteduser")
	}
}

// ---------------------------------------------------------------------------
// Full auth flow integration test
// ---------------------------------------------------------------------------

func TestFullAuthFlow(t *testing.T) {
	deps := newTestDeps(t)
	router := newTestRouter(deps)

	// Step 1: Register a new user.
	regBody := `{"username": "flowuser", "password": "flowpass"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", w.Code, http.StatusCreated)
	}

	// Step 2: Log in to get a token.
	form := url.Values{}
	form.Set("username", "flowuser")
	form.Set("password", "flowpass")

	req = httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("token status = %d, want %d", w.Code, http.StatusOK)
	}

	var tokenResp Token
	if err := json.NewDecoder(w.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}

	// Step 3: Access protected endpoint with the token.
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("protected endpoint status = %d, want %d", w.Code, http.StatusOK)
	}

	var protectedResp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&protectedResp); err != nil {
		t.Fatalf("failed to decode protected response: %v", err)
	}
	if protectedResp["username"] != "flowuser" {
		t.Errorf("protected response username = %q, want %q", protectedResp["username"], "flowuser")
	}
}
