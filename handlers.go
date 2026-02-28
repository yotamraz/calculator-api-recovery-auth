package main

import (
	"net/http"

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

// RegisterHandler handles user registration.
// POST /auth/register
func RegisterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserCreate
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
			return
		}

		// Check for duplicate username
		var existing User
		if err := db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Username already taken"})
			return
		}

		// Hash the password
		hashedPassword, err := HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to hash password"})
			return
		}

		// Create the user
		user := User{
			Username:       req.Username,
			HashedPassword: hashedPassword,
		}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create user"})
			return
		}

		// Return user response with 201 status
		c.JSON(http.StatusCreated, UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		})
	}
}

// TokenHandler handles token issuance via OAuth2 password flow.
// POST /auth/token
// Accepts application/x-www-form-urlencoded with fields: username, password
func TokenHandler(db *gorm.DB, cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		if username == "" || password == "" {
			c.Header("WWW-Authenticate", "Bearer")
			c.JSON(http.StatusUnauthorized, gin.H{
				"detail": "Incorrect username or password",
			})
			return
		}

		// Authenticate the user
		user := AuthenticateUser(db, username, password)
		if user == nil {
			c.Header("WWW-Authenticate", "Bearer")
			c.JSON(http.StatusUnauthorized, gin.H{
				"detail": "Incorrect username or password",
			})
			return
		}

		// Create an access token
		accessToken, err := CreateAccessToken(user.Username, cfg.JWTSecretKey, cfg.AccessTokenExpireMin)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create access token"})
			return
		}

		c.JSON(http.StatusOK, Token{
			AccessToken: accessToken,
			TokenType:   "bearer",
		})
	}
}
