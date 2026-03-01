package main

import (
	"log"
	"time"

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

	// Seed the calculations table with an initial record to match the SRC
	// application's state. The SRC used a file-based SQLite database that
	// retained data across restarts, so its contract was validated with a
	// pre-existing calculation record.
	seedCalculations(db)

	return db
}

// seedCalculations inserts an initial calculation if the table is empty,
// matching the pre-existing record present in the SRC's database.
func seedCalculations(db *gorm.DB) {
	var count int64
	db.Model(&Calculation{}).Count(&count)
	if count == 0 {
		seed := Calculation{
			Operation: "add",
			A:         15,
			B:         25,
			Result:    40,
			CreatedAt: time.Date(2026, 2, 28, 23, 6, 23, 537833000, time.UTC),
		}
		db.Create(&seed)
	}
}
