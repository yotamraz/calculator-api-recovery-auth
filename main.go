package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// initDB initializes the GORM database connection and runs auto-migrations.
func initDB(dsn string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&User{}, &Calculation{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// setupRouter creates and configures the Gin router with all routes.
// This is separated from main() to enable testing with httptest.
func setupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// --- Public Routes ---

	// Health check
	router.GET("/health", healthCheck)

	// Auth endpoints
	auth := router.Group("/auth")
	{
		auth.POST("/register", register(db))
		auth.POST("/token", login(db))
	}

	// --- Protected Routes (require JWT) ---
	protected := router.Group("")
	protected.Use(authMiddleware(db))
	{
		// Calculator operations
		protected.POST("/add", apiAdd)
		protected.POST("/subtract", apiSubtract)
		protected.POST("/multiply", apiMultiply)
		protected.POST("/divide", apiDivide)

		// Calculations CRUD
		protected.POST("/calculations", createCalculation(db))
		protected.GET("/calculations", listCalculations(db))
		protected.GET("/calculations/:id", getCalculation(db))
		protected.DELETE("/calculations/:id", deleteCalculation(db))
	}

	return router
}

func main() {
	// Initialize database
	db := initDB("calculator.db")

	// Setup router
	router := setupRouter(db)

	// Start server on port 8000 (matching FastAPI default)
	log.Println("Starting Calculator API on :8000")
	if err := router.Run(":8000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
