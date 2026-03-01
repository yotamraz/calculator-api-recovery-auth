package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDB initializes a GORM database connection using the SQLite driver
// and runs AutoMigrate for all models. It returns the *gorm.DB instance.
// It drops existing tables first to ensure a clean state on each startup.
func InitDB(databaseURL string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Drop existing tables to ensure clean state on startup.
	migrator := db.Migrator()
	if migrator.HasTable(&Calculation{}) {
		if err := migrator.DropTable(&Calculation{}); err != nil {
			log.Fatalf("Failed to drop Calculation table: %v", err)
		}
	}
	if migrator.HasTable(&User{}) {
		if err := migrator.DropTable(&User{}); err != nil {
			log.Fatalf("Failed to drop User table: %v", err)
		}
	}

	// AutoMigrate creates the tables for User and Calculation.
	if err := db.AutoMigrate(&User{}, &Calculation{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	return db
}
