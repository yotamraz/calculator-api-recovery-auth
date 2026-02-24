package main

import "time"

// --- GORM Models ---

// User is the database model for registered users.
type User struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username       string    `gorm:"uniqueIndex;not null" json:"username"`
	HashedPassword string    `gorm:"not null" json:"-"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// --- Request / Response DTOs ---

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

// LoginForm is the request for the token endpoint (form-encoded).
type LoginForm struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// HealthResponse is the response model for the health check endpoint.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ErrorResponse mirrors FastAPI's error response format.
type ErrorResponse struct {
	Detail string `json:"detail"`
}
