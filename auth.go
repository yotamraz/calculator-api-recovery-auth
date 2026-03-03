package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

// AccessTokenExpireMinutes is the default JWT token lifetime, matching the
// Python implementation's ACCESS_TOKEN_EXPIRE_MINUTES = 30.
const AccessTokenExpireMinutes = 30

// bcryptCost matches Python's bcrypt.gensalt() default cost of 12.
const bcryptCost = 12

// ---------------------------------------------------------------------------
// Context key for authenticated user
// ---------------------------------------------------------------------------

// contextKey is an unexported type used as context key to avoid collisions.
type contextKey string

// userContextKey is the context key for storing the authenticated User.
const userContextKey contextKey = "auth_user"

// UserFromContext extracts the authenticated User from the request context.
// Returns nil if no user is present (should not happen behind AuthMiddleware).
func UserFromContext(ctx context.Context) *User {
	u, _ := ctx.Value(userContextKey).(*User)
	return u
}

// ---------------------------------------------------------------------------
// Password hashing & verification (bcrypt)
// ---------------------------------------------------------------------------

// HashPassword hashes a plain-text password using bcrypt.
// Equivalent to Python's auth.hash_password().
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks a plain-text password against a bcrypt hash.
// Equivalent to Python's auth.verify_password().
func VerifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// ---------------------------------------------------------------------------
// JWT token creation & parsing
// ---------------------------------------------------------------------------

// CreateAccessToken creates an HS256-signed JWT with "sub" and "exp" claims.
// Equivalent to Python's auth.create_access_token().
func CreateAccessToken(username string, secretKey string, expireMinutes int) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub": username,
		"exp": jwt.NewNumericDate(now.Add(time.Duration(expireMinutes) * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ParseToken validates an HS256 JWT and returns the username from the "sub" claim.
// Returns an error if the token is invalid, expired, or missing the "sub" claim.
func ParseToken(tokenString string, secretKey string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		// Ensure signing method is HMAC (HS256).
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", jwt.ErrSignatureInvalid
	}

	return sub, nil
}

// ---------------------------------------------------------------------------
// User authentication helper
// ---------------------------------------------------------------------------

// AuthenticateUser validates credentials and returns the user, or nil if
// invalid. Equivalent to Python's auth.authenticate_user().
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

// ---------------------------------------------------------------------------
// Auth error helper
// ---------------------------------------------------------------------------

// writeAuthError writes a 401 JSON error response with the WWW-Authenticate
// header set to "Bearer", matching the Python implementation's behavior.
func writeAuthError(w http.ResponseWriter, detail string) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	writeError(w, http.StatusUnauthorized, detail)
}

// ---------------------------------------------------------------------------
// Auth middleware
// ---------------------------------------------------------------------------

// AuthMiddleware is a Chi-compatible middleware that validates the JWT Bearer
// token from the Authorization header, loads the full User from the database,
// and stores it in the request context. Returns 401 for missing, invalid, or
// expired tokens. Equivalent to Python's get_current_user dependency.
func (d *Deps) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const credentialsError = "Could not validate credentials"

		// Extract Bearer token from Authorization header.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeAuthError(w, credentialsError)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeAuthError(w, credentialsError)
			return
		}
		tokenString := parts[1]

		// Parse and validate the JWT.
		username, err := ParseToken(tokenString, d.Config.JWTSecretKey)
		if err != nil {
			writeAuthError(w, credentialsError)
			return
		}

		// Look up the user in the database.
		var user User
		if err := d.DB.Where("username = ?", username).First(&user).Error; err != nil {
			writeAuthError(w, credentialsError)
			return
		}

		// Store the full User in the request context and proceed.
		ctx := context.WithValue(r.Context(), userContextKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
