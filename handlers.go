package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

// RegisterHandler returns a Gin handler for user registration.
// POST /auth/register
// Accepts JSON body: {"username": "...", "password": "..."}
// Returns 201 with UserResponse on success.
// Returns 400 with {"detail": "Username already taken"} if the username exists.
func RegisterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserCreate
		if err := c.ShouldBindJSON(&req); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				details := make([]gin.H, 0, len(ve))
				for _, fe := range ve {
					details = append(details, gin.H{
						"loc":  []string{"body", strings.ToLower(fe.Field())},
						"msg":  "Field required",
						"type": "missing",
					})
				}
				c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": details})
				return
			}
			c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": []gin.H{
				{"loc": []string{"body"}, "msg": "Invalid request body", "type": "value_error"},
			}})
			return
		}

		// Check if username already exists
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
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Username already taken"})
			return
		}

		c.JSON(http.StatusCreated, UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		})
	}
}

// LoginHandler returns a Gin handler for user authentication.
// POST /auth/token
// Accepts form-encoded data (username and password fields) to match the
// Python app's OAuth2PasswordRequestForm behavior.
// Returns 200 with Token (access_token + token_type) on success.
// Returns 401 with {"detail": "Incorrect username or password"} on failure.
func LoginHandler(db *gorm.DB, cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		if username == "" || password == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Incorrect username or password"})
			return
		}

		user := AuthenticateUser(db, username, password)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Incorrect username or password"})
			return
		}

		accessToken, err := CreateAccessToken(user.Username, cfg)
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
