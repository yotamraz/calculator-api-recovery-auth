package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// healthCheck handles GET /health.
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: "0.1.0",
	})
}

// registerHandler returns a handler for POST /auth/register.
func registerHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserCreate
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Detail: "Invalid request body",
			})
			return
		}

		// Check for duplicate username.
		var existing User
		result := db.Where("username = ?", req.Username).First(&existing)
		if result.Error == nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Detail: "Username already taken",
			})
			return
		}

		// Hash the password.
		hashed, err := hashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Detail: "Failed to hash password",
			})
			return
		}

		// Create the user.
		user := User{
			Username:       req.Username,
			HashedPassword: hashed,
			CreatedAt:      time.Now().UTC(),
		}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Detail: "Failed to create user",
			})
			return
		}

		c.JSON(http.StatusCreated, UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		})
	}
}

// loginHandler returns a handler for POST /auth/token.
// Accepts application/x-www-form-urlencoded with username and password fields.
func loginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form LoginForm
		if err := c.ShouldBind(&form); err != nil {
			c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Detail: "Invalid request body",
			})
			return
		}

		user := authenticateUser(db, form.Username, form.Password)
		if user == nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Detail: "Incorrect username or password",
			})
			return
		}

		accessToken, err := createAccessToken(user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Detail: "Failed to create access token",
			})
			return
		}

		c.JSON(http.StatusOK, Token{
			AccessToken: accessToken,
			TokenType:   "bearer",
		})
	}
}
