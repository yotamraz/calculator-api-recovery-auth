package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter creates and configures a Gin engine with all routes.
// It accepts a *gorm.DB and Config so that handlers and middleware can access
// the database and configuration (e.g., JWT secret for auth).
func SetupRouter(db *gorm.DB, cfg Config) *gin.Engine {
	r := gin.Default()
	r.HandleMethodNotAllowed = true
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"detail": "Method Not Allowed"})
	})

	// Public routes (no authentication required)
	r.GET("/health", HealthCheck)

	// Public auth routes
	r.POST("/auth/register", RegisterHandler(db))
	r.POST("/auth/token", LoginHandler(db, cfg))

	// Protected routes (authentication middleware guards these endpoints)
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg))
	{
		// Calculator and CRUD routes will be added in milestone 3:
		// protected.POST("/add", ...)
		// protected.POST("/subtract", ...)
		// protected.POST("/multiply", ...)
		// protected.POST("/divide", ...)
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
