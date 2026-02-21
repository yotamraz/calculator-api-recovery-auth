package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ---- Configuration ----

const (
	defaultSecretKey          = "dev-secret-key-change-me-in-production"
	accessTokenExpireMinutes  = 30
	algorithm                 = "HS256"
)

// getSecretKey returns the JWT signing key from the environment or a default.
func getSecretKey() string {
	if key := os.Getenv("JWT_SECRET_KEY"); key != "" {
		return key
	}
	return defaultSecretKey
}

// ---- Password Helpers ----

// hashPassword hashes a plain-text password using bcrypt.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// verifyPassword checks a plain password against a bcrypt hash.
func verifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// ---- JWT Helpers ----

// createAccessToken generates a signed JWT with the given username as subject.
func createAccessToken(username string, expiresDelta time.Duration) (string, error) {
	if expiresDelta == 0 {
		expiresDelta = accessTokenExpireMinutes * time.Minute
	}

	claims := jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresDelta)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(getSecretKey()))
}

// parseToken validates a JWT string and returns the username (sub claim).
func parseToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure we're using the expected signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(getSecretKey()), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	return claims.Subject, nil
}

// ---- User Authentication ----

// authenticateUser validates credentials and returns the user, or nil if invalid.
func authenticateUser(username, password string) *User {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil
	}
	if !verifyPassword(password, user.HashedPassword) {
		return nil
	}
	return &user
}

// ---- Gin Middleware ----

// authMiddleware is a Gin middleware that validates the JWT Bearer token,
// looks up the user in the database, and stores it in the context.
// On failure, it aborts with HTTP 401.
func authMiddleware() gin.HandlerFunc {
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
		username, err := parseToken(tokenString)
		if err != nil || username == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Detail: "Could not validate credentials",
			})
			return
		}

		// Look up the user in the database
		var user User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Detail: "Could not validate credentials",
			})
			return
		}

		// Store user in context for downstream handlers
		c.Set("currentUser", &user)
		c.Next()
	}
}

// getCurrentUser retrieves the authenticated user from the Gin context.
func getCurrentUser(c *gin.Context) *User {
	val, exists := c.Get("currentUser")
	if !exists {
		return nil
	}
	user, ok := val.(*User)
	if !ok {
		return nil
	}
	return user
}
