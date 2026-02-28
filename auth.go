package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ============================================================================
// Password Hashing
// ============================================================================

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// VerifyPassword checks a plaintext password against a bcrypt hash.
func VerifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// ============================================================================
// JWT Token Creation and Parsing
// ============================================================================

// CreateAccessToken creates a signed JWT with HS256 containing the username
// as the "sub" claim and an expiration time.
func CreateAccessToken(username string, secret string, expiryMinutes int) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"exp": jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(expiryMinutes) * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken validates and parses a JWT, returning the username from the "sub" claim.
// Returns an error if the token is invalid, expired, or missing the "sub" claim.
func ParseToken(tokenString string, secret string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	username, ok := claims["sub"].(string)
	if !ok || username == "" {
		return "", errors.New("missing sub claim")
	}

	return username, nil
}

// ============================================================================
// User Authentication Helper
// ============================================================================

// AuthenticateUser looks up a user by username and verifies the password.
// Returns the user if credentials are valid, or nil if not.
func AuthenticateUser(db *gorm.DB, username, password string) *User {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil
	}
	if !VerifyPassword(password, user.HashedPassword) {
		return nil
	}
	return &user
}

// ============================================================================
// JWT Authentication Middleware
// ============================================================================

// AuthMiddleware returns a Gin middleware that validates JWT tokens from the
// Authorization header and injects the authenticated user into the context.
func AuthMiddleware(db *gorm.DB, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"detail": "Could not validate credentials",
			})
			return
		}

		// Extract the Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"detail": "Could not validate credentials",
			})
			return
		}

		tokenString := parts[1]

		// Parse and validate the token
		username, err := ParseToken(tokenString, secret)
		if err != nil {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"detail": "Could not validate credentials",
			})
			return
		}

		// Look up the user in the database
		var user User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"detail": "Could not validate credentials",
			})
			return
		}

		// Store the user in the context for handlers to use
		c.Set("currentUser", &user)
		c.Next()
	}
}

// GetCurrentUser extracts the authenticated user from the Gin context.
// This should only be called within handlers protected by AuthMiddleware.
func GetCurrentUser(c *gin.Context) *User {
	user, exists := c.Get("currentUser")
	if !exists {
		return nil
	}
	return user.(*User)
}
