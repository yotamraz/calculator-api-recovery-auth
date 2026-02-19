package main

import "time"

// ---------------------------------------------------------------------------
// GORM database models
// ---------------------------------------------------------------------------

// User represents a registered user in the database.
type User struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Username       string    `gorm:"uniqueIndex;not null" json:"username"`
	HashedPassword string    `gorm:"not null" json:"-"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Calculation represents a stored calculation in the database.
type Calculation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Operation string    `gorm:"not null" json:"operation"`
	A         float64   `gorm:"not null" json:"a"`
	B         float64   `gorm:"not null" json:"b"`
	Result    float64   `gorm:"not null" json:"result"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// ---------------------------------------------------------------------------
// Request structs
// ---------------------------------------------------------------------------

// UserCreate is the request body for user registration (POST /auth/register).
type UserCreate struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CalculationCreate is the request body for creating a calculation (POST /calculations).
type CalculationCreate struct {
	Operation string  `json:"operation" binding:"required"`
	A         float64 `json:"a" binding:"required"`
	B         float64 `json:"b"`
}

// CalculationRequest is the request body for quick calculator endpoints (/add, /subtract, etc.).
type CalculationRequest struct {
	A float64 `json:"a" binding:"required"`
	B float64 `json:"b"`
}

// ---------------------------------------------------------------------------
// Response structs
// ---------------------------------------------------------------------------

// UserResponse is the response model for user info (POST /auth/register).
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// Token is the response model for an access token (POST /auth/token).
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// CalculationResponse is the response model for a calculation.
type CalculationResponse struct {
	ID        uint      `json:"id"`
	Operation string    `json:"operation"`
	A         float64   `json:"a"`
	B         float64   `json:"b"`
	Result    float64   `json:"result"`
	CreatedAt time.Time `json:"created_at"`
}

// ResultResponse is the response model for quick calculator endpoints.
type ResultResponse struct {
	Result float64 `json:"result"`
}

// HealthResponse is the response model for the health check endpoint.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
