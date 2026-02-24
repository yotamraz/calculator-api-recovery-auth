package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupRouter creates and configures the Gin engine with all routes.
// The db parameter is used by handlers and middleware to access the database.
// This function is also used by tests to create a router with an in-memory DB.
func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// --- Public routes ---
	r.GET("/health", healthCheck)

	// --- Auth routes (public) ---
	auth := r.Group("/auth")
	{
		auth.POST("/register", registerHandler(db))
		auth.POST("/token", loginHandler(db))
	}

	// --- Protected routes (require valid JWT) ---
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db))
	{
		// Calculator endpoints will be added in Milestone 2.
		// Calculation CRUD endpoints will be added in Milestone 2.
	}

	return r
}

func main() {
	// Initialize database.
	db, err := gorm.Open(sqlite.Open("calculator.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models.
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Setup router and start server.
	r := SetupRouter(db)
	if err := r.Run(":8000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
