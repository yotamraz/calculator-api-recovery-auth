package main

import "time"

// ============================================================================
// GORM Database Models
// ============================================================================

// User is the database model for users.
type User struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username       string    `gorm:"uniqueIndex;not null" json:"username"`
	HashedPassword string    `gorm:"not null" json:"-"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Calculation is the database model for stored calculations.
type Calculation struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Operation string    `gorm:"not null" json:"operation"`
	A         float64   `gorm:"not null" json:"a"`
	B         float64   `gorm:"not null" json:"b"`
	Result    float64   `gorm:"not null" json:"result"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// ============================================================================
// API Request/Response DTOs
// ============================================================================

// UserCreate is the request body for user registration.
type UserCreate struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse is the response model for user info.
type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// Token is the response model for an access token.
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// CalculationCreate is the request body for creating a calculation.
type CalculationCreate struct {
	Operation string  `json:"operation" binding:"required"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
}

// CalculationResponse is the response model for a calculation.
type CalculationResponse struct {
	ID        int       `json:"id"`
	Operation string    `json:"operation"`
	A         float64   `json:"a"`
	B         float64   `json:"b"`
	Result    float64   `json:"result"`
	CreatedAt time.Time `json:"created_at"`
}

// CalculationRequest is the request body for calculator endpoints.
type CalculationRequest struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

// ResultResponse is the response body for calculator endpoints.
type ResultResponse struct {
	Result float64 `json:"result"`
}

// HealthResponse is the response model for the health check endpoint.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
