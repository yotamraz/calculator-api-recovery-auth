package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter creates and configures a Gin engine with all routes.
// It accepts a *gorm.DB so that handlers and middleware can access the database.
// The router is structured with route groups to anticipate future middleware
// (e.g., JWT auth) added in later milestones.
func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.HandleMethodNotAllowed = true

	// Public routes (no authentication required)
	r.GET("/health", HealthCheck)

	// Future public auth routes will go here:
	// r.POST("/auth/register", ...)
	// r.POST("/auth/token", ...)

	// Protected routes (authentication middleware will be added in a later milestone)
	// protected := r.Group("/")
	// protected.Use(AuthMiddleware(db))
	// {
	//   protected.POST("/add", ...)
	//   protected.POST("/subtract", ...)
	//   protected.POST("/multiply", ...)
	//   protected.POST("/divide", ...)
	//   protected.POST("/calculations", ...)
	//   protected.GET("/calculations", ...)
	//   protected.GET("/calculations/:id", ...)
	//   protected.DELETE("/calculations/:id", ...)
	// }

	return r
}

func main() {
	cfg := LoadConfig()
	db := InitDB(cfg.DatabaseURL)

	// Store config in a package-level variable for access by auth module (future milestones)
	_ = db

	r := SetupRouter(db)
	r.Run(":8000")
}
