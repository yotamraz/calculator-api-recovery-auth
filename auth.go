package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// --- Password Helpers ---

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword checks a plain password against a bcrypt hash.
func VerifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// --- JWT Helpers ---

// CreateAccessToken creates a signed JWT access token for the given username.
// The token uses HS256 signing with the JWT secret from config and includes
// standard "sub" (subject) and "exp" (expiration) claims matching the Python app.
func CreateAccessToken(username string, cfg Config) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.AccessTokenExpireMin) * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecretKey))
}

// ParseToken validates and decodes a JWT token string. It returns the username
// (from the "sub" claim) if the token is valid, or an error otherwise.
func ParseToken(tokenString string, secretKey string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	if claims.Subject == "" {
		return "", errors.New("missing subject claim")
	}

	return claims.Subject, nil
}

// --- Authentication ---

// AuthenticateUser validates credentials against the database.
// Returns the user if valid, or nil if the username doesn't exist or password is wrong.
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

// --- JWT Auth Middleware ---

// AuthMiddleware returns a Gin middleware that validates JWT Bearer tokens.
// It extracts the token from the Authorization header, decodes it, loads the
// user from the database, and stores it in the Gin context under key "currentUser".
// If any step fails, it aborts with 401 and the message "Could not validate credentials".
func AuthMiddleware(db *gorm.DB, cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		credErr := gin.H{"detail": "Could not validate credentials"}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, credErr)
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, credErr)
			return
		}
		tokenString := parts[1]

		username, err := ParseToken(tokenString, cfg.JWTSecretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, credErr)
			return
		}

		var user User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, credErr)
			return
		}

		// Store the authenticated user in context for handlers to access.
		c.Set("currentUser", &user)
		c.Next()
	}
}

// GetCurrentUser retrieves the authenticated user from the Gin context.
// Returns nil if no user is set (should not happen if AuthMiddleware is applied).
func GetCurrentUser(c *gin.Context) *User {
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
