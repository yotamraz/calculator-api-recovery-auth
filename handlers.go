package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthCheck returns the health status of the service.
// GET /health
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: "0.1.0",
	})
}

// RegisterHandler returns a Gin handler for POST /auth/register.
// It creates a new user with a hashed password after checking for duplicate usernames.
func RegisterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserCreate
		if err := c.ShouldBindJSON(&req); err != nil {
			abortWithValidationError(c, err)
			return
		}

		// Check for duplicate username
		var existing User
		result := db.Where("username = ?", req.Username).First(&existing)
		if result.Error == nil {
			// User already exists
			abortWithDetail(c, http.StatusBadRequest, "Username already taken")
			return
		}

		// Hash password
		hashedPassword, err := HashPassword(req.Password)
		if err != nil {
			abortWithDetail(c, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		// Create user
		user := User{
			Username:       req.Username,
			HashedPassword: hashedPassword,
			CreatedAt:      time.Now().UTC(),
		}
		if err := db.Create(&user).Error; err != nil {
			abortWithDetail(c, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Return the created user
		c.JSON(http.StatusCreated, UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		})
	}
}

// LoginHandler returns a Gin handler for POST /auth/token.
// It accepts form-encoded username and password (matching FastAPI's OAuth2PasswordRequestForm),
// validates credentials, and returns a JWT access token.
func LoginHandler(db *gorm.DB, cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract form fields (application/x-www-form-urlencoded)
		username := c.PostForm("username")
		password := c.PostForm("password")

		// Validate required fields (match FastAPI's 422 for missing fields)
		if username == "" || password == "" {
			abortWithFormValidationError(c, username, password)
			return
		}

		// Authenticate user
		user := AuthenticateUser(db, username, password)
		if user == nil {
			c.Header("WWW-Authenticate", "Bearer")
			abortWithDetail(c, http.StatusUnauthorized, "Incorrect username or password")
			return
		}

		// Create JWT token
		accessToken, err := CreateAccessToken(user.Username, cfg)
		if err != nil {
			abortWithDetail(c, http.StatusInternalServerError, "Failed to create access token")
			return
		}

		c.JSON(http.StatusOK, Token{
			AccessToken: accessToken,
			TokenType:   "bearer",
		})
	}
}
