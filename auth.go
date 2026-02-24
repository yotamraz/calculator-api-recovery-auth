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

// --- Configuration ---

const (
	defaultSecretKey          = "dev-secret-key-change-me-in-production"
	algorithm                 = "HS256"
	accessTokenExpireMinutes  = 30
)

// getJWTSecret returns the JWT signing secret from the environment,
// falling back to a default for local development.
func getJWTSecret() string {
	if secret := os.Getenv("JWT_SECRET_KEY"); secret != "" {
		return secret
	}
	return defaultSecretKey
}

// --- Password Helpers ---

// hashPassword hashes a plaintext password using bcrypt.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// verifyPassword checks a plaintext password against a bcrypt hash.
func verifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// --- JWT Helpers ---

// createAccessToken creates a signed JWT token with the given username as subject.
func createAccessToken(username string, expiresDelta time.Duration) (string, error) {
	if expiresDelta == 0 {
		expiresDelta = accessTokenExpireMinutes * time.Minute
	}

	claims := jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresDelta)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(getJWTSecret()))
}

// validateToken parses and validates a JWT token string, returning the claims.
func validateToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(getJWTSecret()), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// --- Error Helper ---

// jsonError writes a JSON error response matching FastAPI's {"detail": "..."} format.
func jsonError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"detail": message})
}

// --- Gin Auth Middleware ---

// authMiddleware returns a Gin middleware that validates JWT tokens
// and attaches the authenticated user to the request context.
func authMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			jsonError(c, http.StatusUnauthorized, "Not authenticated")
			c.Abort()
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			jsonError(c, http.StatusUnauthorized, "Could not validate credentials")
			c.Abort()
			return
		}
		tokenString := parts[1]

		// Validate token
		claims, err := validateToken(tokenString)
		if err != nil {
			jsonError(c, http.StatusUnauthorized, "Could not validate credentials")
			c.Abort()
			return
		}

		username := claims.Subject
		if username == "" {
			jsonError(c, http.StatusUnauthorized, "Could not validate credentials")
			c.Abort()
			return
		}

		// Look up user in database
		var user User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			jsonError(c, http.StatusUnauthorized, "Could not validate credentials")
			c.Abort()
			return
		}

		// Store user in context for handlers to access
		c.Set("user", user)
		c.Next()
	}
}

// getCurrentUser retrieves the authenticated user from the Gin context.
func getCurrentUser(c *gin.Context) User {
	user, _ := c.Get("user")
	return user.(User)
}
