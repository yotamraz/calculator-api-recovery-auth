package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupTestRouter creates a Gin router backed by an in-memory SQLite
// database, suitable for deterministic integration tests.
func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := SetupDB(":memory:")
	if err != nil {
		t.Fatalf("SetupDB(:memory:) failed: %v", err)
	}

	return SetupRouter(db)
}

// ---------- Health endpoint tests ----------

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /health status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("status = %q, want %q", resp.Status, "ok")
	}
	if resp.Version != "0.1.0" {
		t.Errorf("version = %q, want %q", resp.Version, "0.1.0")
	}
}

func TestHealthEndpointResponseShape(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify exact JSON shape matches FastAPI output.
	var raw map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	// Should contain exactly "status" and "version" keys.
	if len(raw) != 2 {
		t.Errorf("expected 2 keys in response, got %d: %v", len(raw), raw)
	}
	if _, ok := raw["status"]; !ok {
		t.Error("response missing 'status' key")
	}
	if _, ok := raw["version"]; !ok {
		t.Error("response missing 'version' key")
	}
}

func TestHealthEndpointContentType(t *testing.T) {
	router := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	expected := "application/json; charset=utf-8"
	if ct != expected {
		t.Errorf("Content-Type = %q, want %q", ct, expected)
	}
}

func TestHealthEndpointMethodNotAllowed(t *testing.T) {
	router := setupTestRouter(t)

	// POST to /health should not match.
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Errorf("POST /health should not return 200, got %d", w.Code)
	}
}
