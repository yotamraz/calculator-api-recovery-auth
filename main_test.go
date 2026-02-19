package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestRouter creates a Gin router backed by an in-memory SQLite database.
// This helper is reusable by future milestone tests.
func setupTestRouter(t *testing.T) *http.Handler {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	if err := db.AutoMigrate(&User{}, &Calculation{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	router := SetupRouter(db)
	var handler http.Handler = router
	return &handler
}

// setupTestDB creates an in-memory GORM DB and returns both the DB and router.
func setupTestDB(t *testing.T) (*gorm.DB, http.Handler) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	if err := db.AutoMigrate(&User{}, &Calculation{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	router := SetupRouter(db)
	return db, router
}

func TestHealthEndpoint(t *testing.T) {
	_, router := setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("GET /health status = %d, want %d", w.Code, http.StatusOK)
	}

	// Check response body
	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("health status = %q, want %q", resp.Status, "ok")
	}
	if resp.Version != "0.1.0" {
		t.Errorf("health version = %q, want %q", resp.Version, "0.1.0")
	}
}

func TestHealthEndpointResponseShape(t *testing.T) {
	_, router := setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify exact JSON shape matches Python implementation
	var raw map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Should have exactly 2 keys: "status" and "version"
	if len(raw) != 2 {
		t.Errorf("response has %d keys, want 2; got: %v", len(raw), raw)
	}

	if _, ok := raw["status"]; !ok {
		t.Error("response missing 'status' key")
	}
	if _, ok := raw["version"]; !ok {
		t.Error("response missing 'version' key")
	}
}

func TestHealthEndpointContentType(t *testing.T) {
	_, router := setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	want := "application/json; charset=utf-8"
	if ct != want {
		t.Errorf("Content-Type = %q, want %q", ct, want)
	}
}
