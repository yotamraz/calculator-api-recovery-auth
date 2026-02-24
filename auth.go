package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Configuration constants matching the Python implementation.
const (
	defaultSecretKey         = "dev-secret-key-change-me-in-production"
	algorithm                = "HS256"
	accessTokenExpireMinutes = 30
)

// getSecretKey returns the JWT signing key from the environment,
// falling back to a development default.
func getSecretKey() string {
	if key := os.Getenv("JWT_SECRET_KEY"); key != "" {
		return key
	}
	return defaultSecretKey
}

// hashPassword hashes a plain-text password using bcrypt.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// verifyPassword checks a plain password against its bcrypt hash.
func verifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// createAccessToken creates a signed JWT access token with the given subject (username)
// and an expiry of accessTokenExpireMinutes from now.
func createAccessToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"exp": jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(accessTokenExpireMinutes) * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(getSecretKey()))
}

// parseToken validates and parses a JWT token string, returning the claims.
func parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(getSecretKey()), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

// authenticateUser validates credentials and returns the user, or nil if invalid.
func authenticateUser(db *gorm.DB, username, password string) *User {
	var user User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil
	}
	if !verifyPassword(password, user.HashedPassword) {
		return nil
	}
	return &user
}

// AuthMiddleware is a Gin middleware that validates the JWT Bearer token
// and sets the authenticated User in the context.
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Detail: "Not authenticated",
			})
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Detail: "Could not validate credentials",
			})
			return
		}

		tokenString := parts[1]
		claims, err := parseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Detail: "Could not validate credentials",
			})
			return
		}

		username, ok := claims["sub"].(string)
		if !ok || username == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Detail: "Could not validate credentials",
			})
			return
		}

		// Look up the user in the database.
		var user User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Detail: "Could not validate credentials",
			})
			return
		}

		// Store the authenticated user in the context.
		c.Set("user", user)
		c.Next()
	}
}
