package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AppConfig holds the application configuration, accessible to all handlers
// and middleware (e.g., JWT auth in future milestones).
var AppConfig Config

// SetupRouter creates and configures a Gin engine with all routes.
// It accepts a *gorm.DB so that handlers and middleware can access the database.
// The router is structured with route groups to anticipate future middleware
// (e.g., JWT auth) added in later milestones.
func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.HandleMethodNotAllowed = true
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"detail": "Method Not Allowed"})
	})

	// Store the database connection in Gin context for handler access.
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Public routes (no authentication required)
	r.GET("/health", HealthCheck)

	// Auth routes and protected calculator/CRUD routes will be added in later milestones.

	return r
}

func main() {
	cfg := LoadConfig()
	AppConfig = cfg

	db := InitDB(cfg.DatabaseURL)

	r := SetupRouter(db)
	r.Run(":8000")
}
