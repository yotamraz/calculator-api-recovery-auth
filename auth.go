package main

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

// Algorithm used for JWT signing.
const jwtAlgorithm = "HS256"

// Default token expiry duration (matches Python ACCESS_TOKEN_EXPIRE_MINUTES).
const accessTokenExpireMinutes = 30

// ---------------------------------------------------------------------------
// Context key for authenticated user
// ---------------------------------------------------------------------------

// contextKey is an unexported type used for context keys to avoid collisions.
type contextKey string

// userContextKey stores the authenticated user ID in request context.
const userContextKey contextKey = "userID"

// ---------------------------------------------------------------------------
// JWT helpers
// ---------------------------------------------------------------------------

// createAccessToken generates a signed HS256 JWT for the given username.
func createAccessToken(username string, secret string, expiry time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub": username,
		"exp": jwt.NewNumericDate(now.Add(expiry)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// parseAccessToken validates and parses a JWT, returning the claims.
func parseAccessToken(tokenString string, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwtAlgorithm}))
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

// ---------------------------------------------------------------------------
// Password hashing helpers
// ---------------------------------------------------------------------------

// hashPassword hashes a plain-text password using bcrypt.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// verifyPassword checks a plain-text password against a bcrypt hash.
func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ---------------------------------------------------------------------------
// Auth middleware (placeholder — full wiring in Milestone 2)
// ---------------------------------------------------------------------------

// AuthMiddleware validates the Bearer token and injects the user ID into the
// request context. Full implementation will be completed in Milestone 2.
func (d *Deps) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Bearer token from Authorization header.
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			writeError(w, http.StatusUnauthorized, "Could not validate credentials")
			return
		}
		tokenString := authHeader[7:]

		// Parse and validate the token.
		claims, err := parseAccessToken(tokenString, d.Config.JWTSecretKey)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Could not validate credentials")
			return
		}

		// Extract username from claims.
		username, ok := claims["sub"].(string)
		if !ok || username == "" {
			writeError(w, http.StatusUnauthorized, "Could not validate credentials")
			return
		}

		// Look up the user in the database.
		var user User
		if err := d.DB.Where("username = ?", username).First(&user).Error; err != nil {
			writeError(w, http.StatusUnauthorized, "Could not validate credentials")
			return
		}

		// Store user ID in request context.
		ctx := context.WithValue(r.Context(), userContextKey, user.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// userIDFromContext retrieves the authenticated user ID from the request context.
func userIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userContextKey).(int)
	return id, ok
}
