package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter creates and configures a Gin engine with all routes.
// It accepts a *gorm.DB and Config so that handlers and middleware can access
// the database and configuration (e.g., JWT secret).
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

	// Protected routes (authentication required via JWT middleware)
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg))
	{
		// Calculator endpoints
		protected.POST("/add", AddHandler())
		protected.POST("/subtract", SubtractHandler())
		protected.POST("/multiply", MultiplyHandler())
		protected.POST("/divide", DivideHandler())

		// Calculation CRUD endpoints
		protected.POST("/calculations", CreateCalculationHandler(db))
		protected.GET("/calculations", ListCalculationsHandler(db))
		protected.GET("/calculations/:id", GetCalculationHandler(db))
		protected.DELETE("/calculations/:id", DeleteCalculationHandler(db))
	}

	return r
}

func main() {
	cfg := LoadConfig()
	db := InitDB(cfg.DatabaseURL)

	r := SetupRouter(db, cfg)
	r.Run(":8000")
}
