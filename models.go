package main

import "time"

// ---- GORM Database Models ----

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

// ---- Request Structs ----

// UserCreate is the request body for user registration.
type UserCreate struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginForm is the request body for the token endpoint (form-encoded).
type LoginForm struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// CalculationRequest is the request body for calculator endpoints (/add, /subtract, etc.).
type CalculationRequest struct {
	A float64 `json:"a" binding:"required"`
	B float64 `json:"b" binding:"required"`
}

// CalculationCreate is the request body for creating a stored calculation.
type CalculationCreate struct {
	Operation string  `json:"operation" binding:"required"`
	A         float64 `json:"a" binding:"required"`
	B         float64 `json:"b" binding:"required"`
}

// ---- Response Structs ----

// UserResponse is the response model for user info.
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// Token is the response model for an access token.
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// ResultResponse is the response body for calculator endpoints.
type ResultResponse struct {
	Result float64 `json:"result"`
}

// CalculationResponse is the response model for a stored calculation.
type CalculationResponse struct {
	ID        uint      `json:"id"`
	Operation string    `json:"operation"`
	A         float64   `json:"a"`
	B         float64   `json:"b"`
	Result    float64   `json:"result"`
	CreatedAt time.Time `json:"created_at"`
}

// HealthResponse is the response model for the health check endpoint.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ErrorResponse is a standard error response matching FastAPI's {"detail": "..."} format.
type ErrorResponse struct {
	Detail string `json:"detail"`
}

// ---- Operation Map ----

// operationFunc is a calculator function that takes two float64s and returns a float64.
// The divide operation is a special case handled separately due to its error return.
type operationFunc func(float64, float64) float64

// OPERATIONS maps operation names to their corresponding core functions.
// Note: "div" is handled separately in handlers because divide() returns an error.
var OPERATIONS = map[string]operationFunc{
	"add": add,
	"sub": subtract,
	"mul": multiply,
}
