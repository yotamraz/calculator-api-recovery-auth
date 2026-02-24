package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ---------- Response / Request structs ----------

// HealthResponse is the JSON shape returned by GET /health.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ErrorResponse mirrors FastAPI's HTTPException body: {"detail": "..."}.
type ErrorResponse struct {
	Detail string `json:"detail"`
}

// ---------- Database setup ----------

// SetupDB opens a GORM SQLite connection using the given DSN and
// auto-migrates the provided models. Pass ":memory:" for tests or a
// file path (e.g. "calculator.db") for production.
func SetupDB(dsn string, models ...interface{}) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			return nil, err
		}
	}

	return db, nil
}
