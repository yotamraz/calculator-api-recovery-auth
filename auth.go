package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// contextKeyUser is the key used to store/retrieve the authenticated user in the Gin context.
const contextKeyUser = "authenticated_user"

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword checks a plain password against its bcrypt hash.
func VerifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// CreateAccessToken creates a signed JWT access token with the given username as the "sub" claim.
func CreateAccessToken(username string, cfg Config) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.AccessTokenExpireMin) * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecretKey))
}

// AuthenticateUser validates credentials and returns the user, or nil if invalid.
func AuthenticateUser(db *gorm.DB, username, password string) *User {
	var user User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil
	}
	if !VerifyPassword(password, user.HashedPassword) {
		return nil
	}
	return &user
}

// SetCurrentUser stores the authenticated user in the Gin context.
func SetCurrentUser(c *gin.Context, user *User) {
	c.Set(contextKeyUser, user)
}

// GetCurrentUser retrieves the authenticated user from the Gin context.
// Returns nil if no user is set.
func GetCurrentUser(c *gin.Context) *User {
	val, exists := c.Get(contextKeyUser)
	if !exists {
		return nil
	}
	user, ok := val.(*User)
	if !ok {
		return nil
	}
	return user
}

// abortWithDetail is a helper that produces FastAPI-compatible error responses
// with the shape {"detail": "<message>"}.
func abortWithDetail(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"detail": message})
}

// validationErrorItem represents a single FastAPI-style validation error.
type validationErrorItem struct {
	Loc  []string `json:"loc"`
	Msg  string   `json:"msg"`
	Type string   `json:"type"`
}

// abortWithValidationError produces a FastAPI-compatible 422 response with
// detail as an array of validation error objects, parsed from Gin binding errors.
func abortWithValidationError(c *gin.Context, err error) {
	var details []validationErrorItem

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			details = append(details, validationErrorItem{
				Loc:  []string{"body", strings.ToLower(fe.Field())},
				Msg:  "Field required",
				Type: "missing",
			})
		}
	} else {
		// Fallback for non-validation binding errors (e.g. malformed JSON)
		details = append(details, validationErrorItem{
			Loc:  []string{"body"},
			Msg:  "Invalid request body",
			Type: "value_error",
		})
	}

	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"detail": details})
}

// abortWithFormValidationError produces a FastAPI-compatible 422 response for
// missing form fields (username/password on the token endpoint).
func abortWithFormValidationError(c *gin.Context, username, password string) {
	var details []validationErrorItem
	if username == "" {
		details = append(details, validationErrorItem{
			Loc:  []string{"body", "username"},
			Msg:  "Field required",
			Type: "missing",
		})
	}
	if password == "" {
		details = append(details, validationErrorItem{
			Loc:  []string{"body", "password"},
			Msg:  "Field required",
			Type: "missing",
		})
	}
	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"detail": details})
}

// AuthMiddleware returns a Gin middleware that validates JWT bearer tokens
// and injects the authenticated user into the request context.
// On failure, it aborts with 401 and a WWW-Authenticate: Bearer header,
// matching the Python FastAPI app's error responses.
func AuthMiddleware(db *gorm.DB, cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		credentialsError := "Could not validate credentials"

		// Extract the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Header("WWW-Authenticate", "Bearer")
			abortWithDetail(c, http.StatusUnauthorized, credentialsError)
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Header("WWW-Authenticate", "Bearer")
			abortWithDetail(c, http.StatusUnauthorized, credentialsError)
			return
		}
		tokenString := parts[1]

		// Parse and validate the JWT
		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JWTSecretKey), nil
		})
		if err != nil || !token.Valid {
			c.Header("WWW-Authenticate", "Bearer")
			abortWithDetail(c, http.StatusUnauthorized, credentialsError)
			return
		}

		// Extract the username from the "sub" claim
		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok {
			c.Header("WWW-Authenticate", "Bearer")
			abortWithDetail(c, http.StatusUnauthorized, credentialsError)
			return
		}

		username := claims.Subject
		if username == "" {
			c.Header("WWW-Authenticate", "Bearer")
			abortWithDetail(c, http.StatusUnauthorized, credentialsError)
			return
		}

		// Look up the user in the database
		var user User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			c.Header("WWW-Authenticate", "Bearer")
			abortWithDetail(c, http.StatusUnauthorized, credentialsError)
			return
		}

		// Store user in context and proceed
		SetCurrentUser(c, &user)
		c.Next()
	}
}
