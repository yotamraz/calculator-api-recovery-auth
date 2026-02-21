package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// db is the package-level database connection used by all handlers.
var db *gorm.DB

// initDB initializes the database connection and runs auto-migrations.
func initDB(dsn string) (*gorm.DB, error) {
	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate models
	if err := database.AutoMigrate(&User{}, &Calculation{}); err != nil {
		return nil, err
	}

	return database, nil
}

// SetupRouter creates and configures the Gin router with all routes and middleware.
// This function is exported for use in tests.
func SetupRouter(database *gorm.DB) *gin.Engine {
	db = database

	router := gin.Default()

	// Health check (public)
	router.GET("/health", healthCheck)

	// Auth endpoints (public)
	auth := router.Group("/auth")
	{
		auth.POST("/register", register)
		auth.POST("/token", login)
	}

	// Protected endpoints (require JWT)
	protected := router.Group("/")
	protected.Use(authMiddleware())
	{
		// Calculator operations
		protected.POST("/add", apiAdd)
		protected.POST("/subtract", apiSubtract)
		protected.POST("/multiply", apiMultiply)
		protected.POST("/divide", apiDivide)

		// Calculations CRUD
		protected.POST("/calculations", createCalculation)
		protected.GET("/calculations", listCalculations)
		protected.GET("/calculations/:id", getCalculation)
		protected.DELETE("/calculations/:id", deleteCalculation)
	}

	return router
}

func main() {
	// Initialize database
	database, err := initDB("calculator.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup router
	router := SetupRouter(database)

	// Start server on port 8000 (matching FastAPI default)
	log.Println("Starting Calculator API server on :8000")
	if err := router.Run(":8000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
