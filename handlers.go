package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// ValidationErrorDetail represents a single validation error in FastAPI/Pydantic v2 format.
type ValidationErrorDetail struct {
	Loc   []string    `json:"loc"`
	Msg   string      `json:"msg"`
	Type  string      `json:"type"`
	Input interface{} `json:"input"`
}

// formatValidationErrors converts Gin/validator errors into FastAPI-style validation error array.
func formatValidationErrors(err error, input interface{}) []ValidationErrorDetail {
	var details []ValidationErrorDetail
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			fieldName := strings.ToLower(fe.Field())
			details = append(details, ValidationErrorDetail{
				Loc:   []string{"body", fieldName},
				Msg:   "Field required",
				Type:  "missing",
				Input: input,
			})
		}
	} else {
		details = append(details, ValidationErrorDetail{
			Loc:   []string{"body"},
			Msg:   err.Error(),
			Type:  "value_error",
			Input: input,
		})
	}
	return details
}

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
		// Read body first so we can include it as "input" in validation errors
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse raw body as generic JSON for the input field
		var rawInput interface{}
		json.Unmarshal(bodyBytes, &rawInput)

		var req UserCreate
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": formatValidationErrors(err, rawInput)})
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

// TokenRequest is the request body for the token endpoint.
// Supports both form-encoded and JSON formats.
type TokenRequest struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

// TokenHandler handles token issuance via OAuth2 password flow.
// POST /auth/token
// Accepts application/x-www-form-urlencoded or JSON with fields: username, password
func TokenHandler(db *gorm.DB, cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TokenRequest
		if err := c.ShouldBind(&req); err != nil {
			c.Header("WWW-Authenticate", "Bearer")
			c.JSON(http.StatusUnauthorized, gin.H{
				"detail": "Incorrect username or password",
			})
			return
		}

		username := req.Username
		password := req.Password

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
