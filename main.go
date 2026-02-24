package main

import (
	"log"
	"os"
)

func main() {
	// --- Configuration ---
	dbDSN := "calculator.db"
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		dbDSN = v
	}

	port := ":8000"
	if v := os.Getenv("PORT"); v != "" {
		port = ":" + v
	}

	// --- Database ---
	db, err := SetupDB(dbDSN)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// --- Router ---
	r := SetupRouter(db)

	// --- Start server ---
	log.Printf("Starting server on %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
