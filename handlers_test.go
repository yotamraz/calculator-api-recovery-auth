package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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
