package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CalcFunc is a unified function signature for calculator operations.
// All operations return (float64, error) so they can be stored in a dispatch map.
type CalcFunc func(float64, float64) (float64, error)

// OPERATIONS maps operation name strings to their corresponding calculator functions,
// mirroring the Python app's OPERATIONS dict.
var OPERATIONS = map[string]CalcFunc{
	"add": func(a, b float64) (float64, error) { return Add(a, b), nil },
	"sub": func(a, b float64) (float64, error) { return Subtract(a, b), nil },
	"mul": func(a, b float64) (float64, error) { return Multiply(a, b), nil },
	"div": Divide,
}

// HealthCheck returns the health status of the service.
// GET /health
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: "0.1.0",
	})
}

// RegisterHandler returns a Gin handler for POST /auth/register.
// It creates a new user with a hashed password after checking for duplicate usernames.
func RegisterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read raw body for potential validation error responses
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		var rawBody map[string]interface{}
		json.Unmarshal(bodyBytes, &rawBody)

		// Mask sensitive fields in the raw body for validation error responses
		if rawBody != nil {
			if _, hasPassword := rawBody["password"]; hasPassword {
				rawBody["password"] = "********"
			}
		}

		var req UserCreate
		if err := c.ShouldBindJSON(&req); err != nil {
			abortWithValidationError(c, err, rawBody)
			return
		}

		// Check for duplicate username
		var existing User
		result := db.Where("username = ?", req.Username).First(&existing)
		if result.Error == nil {
			// User already exists
			abortWithDetail(c, http.StatusBadRequest, "Username already taken")
			return
		}

		// Hash password
		hashedPassword, err := HashPassword(req.Password)
		if err != nil {
			abortWithDetail(c, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		// Create user
		user := User{
			Username:       req.Username,
			HashedPassword: hashedPassword,
			CreatedAt:      time.Now().UTC(),
		}
		if err := db.Create(&user).Error; err != nil {
			abortWithDetail(c, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Return the created user
		c.JSON(http.StatusCreated, UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		})
	}
}

// LoginHandler returns a Gin handler for POST /auth/token.
// It accepts form-encoded username and password (matching FastAPI's OAuth2PasswordRequestForm),
// validates credentials, and returns a JWT access token.
func LoginHandler(db *gorm.DB, cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract form fields (application/x-www-form-urlencoded)
		username := c.PostForm("username")
		password := c.PostForm("password")

		// Validate required fields (match FastAPI's 422 for missing fields)
		if username == "" || password == "" {
			abortWithFormValidationError(c, username, password)
			return
		}

		// Authenticate user
		user := AuthenticateUser(db, username, password)
		if user == nil {
			c.Header("WWW-Authenticate", "Bearer")
			abortWithDetail(c, http.StatusUnauthorized, "Incorrect username or password")
			return
		}

		// Create JWT token
		accessToken, err := CreateAccessToken(user.Username, cfg)
		if err != nil {
			abortWithDetail(c, http.StatusInternalServerError, "Failed to create access token")
			return
		}

		c.JSON(http.StatusOK, Token{
			AccessToken: accessToken,
			TokenType:   "bearer",
		})
	}
}

// --- Calculator Endpoints ---

// AddHandler returns a Gin handler for POST /add.
// It adds two numbers and returns the result.
func AddHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CalculationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			abortWithDetail(c, http.StatusUnprocessableEntity, "Invalid request body")
			return
		}
		c.JSON(http.StatusOK, ResultResponse{Result: Add(req.A, req.B)})
	}
}

// SubtractHandler returns a Gin handler for POST /subtract.
// It subtracts b from a and returns the result.
func SubtractHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CalculationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			abortWithDetail(c, http.StatusUnprocessableEntity, "Invalid request body")
			return
		}
		c.JSON(http.StatusOK, ResultResponse{Result: Subtract(req.A, req.B)})
	}
}

// MultiplyHandler returns a Gin handler for POST /multiply.
// It multiplies two numbers and returns the result.
func MultiplyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CalculationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			abortWithDetail(c, http.StatusUnprocessableEntity, "Invalid request body")
			return
		}
		c.JSON(http.StatusOK, ResultResponse{Result: Multiply(req.A, req.B)})
	}
}

// DivideHandler returns a Gin handler for POST /divide.
// It divides a by b and returns the result, or 400 "Cannot divide by zero".
func DivideHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CalculationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			abortWithDetail(c, http.StatusUnprocessableEntity, "Invalid request body")
			return
		}
		result, err := Divide(req.A, req.B)
		if err != nil {
			abortWithDetail(c, http.StatusBadRequest, err.Error())
			return
		}
		c.JSON(http.StatusOK, ResultResponse{Result: result})
	}
}

// --- CRUD Endpoints for Calculations ---

// CreateCalculationHandler returns a Gin handler for POST /calculations.
// It creates and stores a new calculation by dispatching to the appropriate core function.
func CreateCalculationHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CalculationCreate
		if err := c.ShouldBindJSON(&req); err != nil {
			abortWithDetail(c, http.StatusUnprocessableEntity, "Invalid request body")
			return
		}

		opFunc, ok := OPERATIONS[req.Operation]
		if !ok {
			msg := fmt.Sprintf("Unknown operation: %s. Use: ['add', 'sub', 'mul', 'div']", req.Operation)
			abortWithDetail(c, http.StatusBadRequest, msg)
			return
		}

		result, err := opFunc(req.A, req.B)
		if err != nil {
			abortWithDetail(c, http.StatusBadRequest, err.Error())
			return
		}

		calculation := Calculation{
			Operation: req.Operation,
			A:         req.A,
			B:         req.B,
			Result:    result,
			CreatedAt: time.Now().UTC(),
		}
		if err := db.Create(&calculation).Error; err != nil {
			abortWithDetail(c, http.StatusInternalServerError, "Failed to save calculation")
			return
		}

		c.JSON(http.StatusCreated, CalculationResponse{
			ID:        calculation.ID,
			Operation: calculation.Operation,
			A:         calculation.A,
			B:         calculation.B,
			Result:    calculation.Result,
			CreatedAt: calculation.CreatedAt,
		})
	}
}

// ListCalculationsHandler returns a Gin handler for GET /calculations.
// It lists all stored calculations ordered by created_at descending.
func ListCalculationsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var calculations []Calculation
		db.Order("created_at desc").Find(&calculations)

		resp := make([]CalculationResponse, len(calculations))
		for i, calc := range calculations {
			resp[i] = CalculationResponse{
				ID:        calc.ID,
				Operation: calc.Operation,
				A:         calc.A,
				B:         calc.B,
				Result:    calc.Result,
				CreatedAt: calc.CreatedAt,
			}
		}

		c.JSON(http.StatusOK, resp)
	}
}

// GetCalculationHandler returns a Gin handler for GET /calculations/:id.
// It retrieves a specific calculation by ID, returning 404 if not found.
func GetCalculationHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			abortWithDetail(c, http.StatusNotFound, "Calculation not found")
			return
		}

		var calculation Calculation
		result := db.First(&calculation, id)
		if result.Error != nil {
			abortWithDetail(c, http.StatusNotFound, "Calculation not found")
			return
		}

		c.JSON(http.StatusOK, CalculationResponse{
			ID:        calculation.ID,
			Operation: calculation.Operation,
			A:         calculation.A,
			B:         calculation.B,
			Result:    calculation.Result,
			CreatedAt: calculation.CreatedAt,
		})
	}
}

// DeleteCalculationHandler returns a Gin handler for DELETE /calculations/:id.
// It deletes a calculation by ID, returning 204 on success or 404 if not found.
func DeleteCalculationHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			abortWithDetail(c, http.StatusNotFound, "Calculation not found")
			return
		}

		var calculation Calculation
		result := db.First(&calculation, id)
		if result.Error != nil {
			abortWithDetail(c, http.StatusNotFound, "Calculation not found")
			return
		}

		db.Delete(&calculation)
		c.Status(http.StatusNoContent)
	}
}
