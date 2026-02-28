package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDB initializes a GORM database connection using the SQLite driver
// and runs AutoMigrate for all models. It returns the *gorm.DB instance.
func InitDB(databaseURL string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// AutoMigrate creates or updates the tables for User and Calculation.
	if err := db.AutoMigrate(&User{}, &Calculation{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	return db
}
