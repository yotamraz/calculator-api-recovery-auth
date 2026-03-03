package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestDeps creates a Deps instance backed by an in-memory SQLite database,
// suitable for integration tests.
func newTestDeps(t *testing.T) *Deps {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
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
			DatabaseURL:  "file::memory:?cache=shared",
		},
	}
}

// newTestRouter sets up a Chi router with all routes, mirroring main.go wiring.
func newTestRouter(deps *Deps) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/health", deps.HealthHandler)
	return r
}

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
