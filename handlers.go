package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// appVersion is the version string returned by the health endpoint,
// matching the Python FastAPI app's version ("0.1.0").
const appVersion = "0.1.0"

// abortWithError writes a JSON error response in the FastAPI style
// {"detail": "..."} and aborts the Gin context.
func abortWithError(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(code, ErrorResponse{Detail: msg})
}

// ---------- Handlers ----------

// healthHandler responds to GET /health.
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: appVersion,
	})
}

// ---------- Router setup ----------

// SetupRouter creates and returns a fully configured Gin engine.
// It accepts a GORM database handle so that handlers needing persistence
// can access it via the "db" key in the Gin context.  Callers (main and
// tests) provide either a file-backed or in-memory SQLite database.
func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default() // includes Logger and Recovery middleware

	// Store the DB in Gin's context so handlers can retrieve it.
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// --- Public routes ---
	r.GET("/health", healthHandler)

	return r
}
