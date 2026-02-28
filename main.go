package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter creates and configures a Gin engine with all routes.
// It accepts a *gorm.DB and Config so that handlers and middleware can
// access the database and configuration.
func SetupRouter(db *gorm.DB, cfg Config) *gin.Engine {
	r := gin.Default()
	r.HandleMethodNotAllowed = true
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"detail": "Method Not Allowed"})
	})

	// Public routes (no authentication required)
	r.GET("/health", HealthCheck)

	// Auth routes (public)
	r.POST("/auth/register", RegisterHandler(db))
	r.POST("/auth/token", TokenHandler(db, cfg))

	// Protected routes (authentication middleware applied)
	// Calculator and CRUD endpoints will be added in Milestone 3.
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg.JWTSecretKey))
	{
		// Calculator endpoints (to be implemented in Milestone 3)
		// protected.POST("/add", ...)
		// protected.POST("/subtract", ...)
		// protected.POST("/multiply", ...)
		// protected.POST("/divide", ...)

		// Calculation CRUD endpoints (to be implemented in Milestone 3)
		// protected.POST("/calculations", ...)
		// protected.GET("/calculations", ...)
		// protected.GET("/calculations/:id", ...)
		// protected.DELETE("/calculations/:id", ...)
	}

	return r
}

func main() {
	cfg := LoadConfig()
	db := InitDB(cfg.DatabaseURL)

	r := SetupRouter(db, cfg)
	r.Run(":8000")
}
