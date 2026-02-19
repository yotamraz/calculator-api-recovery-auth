package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const appVersion = "0.1.0"

// SetupRouter creates and configures a Gin engine with all routes.
// It accepts a *gorm.DB so that tests can inject an in-memory SQLite database.
func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Store the DB handle in Gin's context via middleware so handlers can access it.
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// --- Health check (no auth required) ---
	r.GET("/health", healthHandler)

	return r
}

// healthHandler returns the service health status and version.
func healthHandler(c *gin.Context) {
	c.JSON(200, HealthResponse{
		Status:  "ok",
		Version: appVersion,
	})
}

// initDB opens a GORM connection to the given SQLite DSN and auto-migrates models.
func initDB(dsn string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Auto-migrate creates the tables for User and Calculation if they don't exist.
	if err := db.AutoMigrate(&User{}, &Calculation{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func main() {
	// Database DSN — defaults to file-based SQLite.
	dsn := "calculator.db"
	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		dsn = envDSN
	}

	db := initDB(dsn)
	r := SetupRouter(db)

	// Listen on port 8000 to match the original FastAPI default.
	port := ":8000"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	log.Printf("Starting Calculator API %s on %s", appVersion, port)
	if err := r.Run(port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
