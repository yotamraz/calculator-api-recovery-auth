package main

import (
	"os"
	"time"

	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// Config holds application configuration loaded from environment variables.
type Config struct {
	JWTSecretKey string
	ServerAddr   string
	DatabaseURL  string
}

// LoadConfig reads configuration from environment variables with sensible
// development defaults.
func LoadConfig() Config {
	cfg := Config{
		JWTSecretKey: "dev-secret-key-change-me-in-production",
		ServerAddr:   ":8000",
		DatabaseURL:  "calculator.db",
	}
	if v := os.Getenv("JWT_SECRET_KEY"); v != "" {
		cfg.JWTSecretKey = v
	}
	if v := os.Getenv("SERVER_ADDR"); v != "" {
		cfg.ServerAddr = v
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.DatabaseURL = v
	}
	return cfg
}

// ---------------------------------------------------------------------------
// Dependency Injection
// ---------------------------------------------------------------------------

// Deps holds shared dependencies (database, config) injected into handlers.
type Deps struct {
	DB     *gorm.DB
	Config Config
}

// ---------------------------------------------------------------------------
// Database Models (GORM)
// ---------------------------------------------------------------------------

// User represents a registered user in the database.
type User struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username       string    `gorm:"uniqueIndex;not null" json:"username"`
	HashedPassword string    `gorm:"not null" json:"hashed_password"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Calculation represents a stored calculation in the database.
type Calculation struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Operation string    `gorm:"not null" json:"operation"`
	A         float64   `gorm:"not null" json:"a"`
	B         float64   `gorm:"not null" json:"b"`
	Result    float64   `gorm:"not null" json:"result"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

// HealthResponse is the response body for GET /health.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// CalculationRequest is the request body for calculator operation endpoints
// (POST /add, /subtract, /multiply, /divide).
type CalculationRequest struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

// ResultResponse is the response body for calculator operation endpoints.
type ResultResponse struct {
	Result float64 `json:"result"`
}

// CalculationCreateRequest is the request body for POST /calculations.
type CalculationCreateRequest struct {
	Operation string  `json:"operation"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
}

// CalculationResponse is the response body for calculation CRUD endpoints.
type CalculationResponse struct {
	ID        int       `json:"id"`
	Operation string    `json:"operation"`
	A         float64   `json:"a"`
	B         float64   `json:"b"`
	Result    float64   `json:"result"`
	CreatedAt time.Time `json:"created_at"`
}

// UserCreateRequest is the request body for POST /auth/register.
type UserCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserResponse is the response body for user-related endpoints.
type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// Token is the response body for POST /auth/token.
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// ErrorResponse is the standard error response body matching FastAPI's
// {"detail": "..."} format.
type ErrorResponse struct {
	Detail string `json:"detail"`
}

// ValidationErrorItem represents a single validation error in the FastAPI format.
type ValidationErrorItem struct {
	Loc  []string `json:"loc"`
	Msg  string   `json:"msg"`
	Type string   `json:"type"`
}

// ValidationErrorResponse matches FastAPI's 422 validation error format
// where "detail" is an array of validation error items.
type ValidationErrorResponse struct {
	Detail []ValidationErrorItem `json:"detail"`
}
