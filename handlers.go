package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ---- Health Check ----

// healthCheck handles GET /health.
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: "0.1.0",
	})
}

// ---- Validation Error Helpers ----

// fieldJSONName returns the JSON tag name for a struct field, falling back to lowercase.
func fieldJSONName(field string) string {
	// Map Go struct field names to their JSON tag equivalents
	fieldMap := map[string]string{
		"Username":  "username",
		"Password":  "password",
		"Operation": "operation",
		"A":         "a",
		"B":         "b",
	}
	if name, ok := fieldMap[field]; ok {
		return name
	}
	return strings.ToLower(field)
}

// formatValidationErrors converts Gin/validator errors to FastAPI-compatible format.
func formatValidationErrors(err error, rawBody json.RawMessage) interface{} {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		var rawInput interface{}
		json.Unmarshal(rawBody, &rawInput)

		details := make([]map[string]interface{}, 0, len(ve))
		for _, fe := range ve {
			errType := "value_error"
			msg := "Invalid value"
			if fe.Tag() == "required" {
				errType = "missing"
				msg = "Field required"
			}
			detail := map[string]interface{}{
				"type":  errType,
				"loc":   []string{"body", fieldJSONName(fe.Field())},
				"msg":   msg,
				"input": rawInput,
			}
			details = append(details, detail)
		}
		return gin.H{"detail": details}
	}
	return gin.H{"detail": "Invalid request body"}
}

// bindJSONWithValidation binds JSON and returns FastAPI-compatible errors on failure.
// Returns (bodyBytes, error). If error is non-nil, the error response has already been sent.
func bindJSONWithValidation(c *gin.Context, obj interface{}) (json.RawMessage, bool) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "Invalid request body"})
		return nil, false
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(http.StatusUnprocessableEntity, formatValidationErrors(err, bodyBytes))
		return nil, false
	}
	return bodyBytes, true
}

// ---- Auth Endpoints ----

// register handles POST /auth/register.
func register(c *gin.Context) {
	var req UserCreate
	if _, ok := bindJSONWithValidation(c, &req); !ok {
		return
	}

	// Check for duplicate username
	var existing User
	if err := db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Detail: "Username already taken"})
		return
	}

	// Hash password
	hashed, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Detail: "Failed to hash password"})
		return
	}

	user := User{
		Username:       req.Username,
		HashedPassword: hashed,
		CreatedAt:      time.Now().UTC(),
	}
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Detail: "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	})
}

// login handles POST /auth/token.
// Accepts form-encoded username and password (OAuth2 password grant style).
func login(c *gin.Context) {
	var form LoginForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Detail: "Invalid request"})
		return
	}

	user := authenticateUser(form.Username, form.Password)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Detail: "Incorrect username or password"})
		return
	}

	token, err := createAccessToken(user.Username, accessTokenExpireMinutes*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Detail: "Failed to create token"})
		return
	}

	c.JSON(http.StatusOK, Token{
		AccessToken: token,
		TokenType:   "bearer",
	})
}

// ---- Calculator Endpoints ----

// apiAdd handles POST /add.
func apiAdd(c *gin.Context) {
	var req CalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Detail: "Invalid request body"})
		return
	}
	c.JSON(http.StatusOK, ResultResponse{Result: add(req.A, req.B)})
}

// apiSubtract handles POST /subtract.
func apiSubtract(c *gin.Context) {
	var req CalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Detail: "Invalid request body"})
		return
	}
	c.JSON(http.StatusOK, ResultResponse{Result: subtract(req.A, req.B)})
}

// apiMultiply handles POST /multiply.
func apiMultiply(c *gin.Context) {
	var req CalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Detail: "Invalid request body"})
		return
	}
	c.JSON(http.StatusOK, ResultResponse{Result: multiply(req.A, req.B)})
}

// apiDivide handles POST /divide.
func apiDivide(c *gin.Context) {
	var req CalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Detail: "Invalid request body"})
		return
	}
	result, err := divide(req.A, req.B)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Detail: err.Error()})
		return
	}
	c.JSON(http.StatusOK, ResultResponse{Result: result})
}

// ---- CRUD Endpoints for Calculations ----

// createCalculation handles POST /calculations.
func createCalculation(c *gin.Context) {
	var req CalculationCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Detail: "Invalid request body"})
		return
	}

	// Compute the result
	var result float64
	if req.Operation == "div" {
		r, err := divide(req.A, req.B)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Detail: err.Error()})
			return
		}
		result = r
	} else if fn, ok := OPERATIONS[req.Operation]; ok {
		result = fn(req.A, req.B)
	} else {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Detail: "Unknown operation: " + req.Operation + ". Use: ['add', 'sub', 'mul', 'div']",
		})
		return
	}

	calc := Calculation{
		Operation: req.Operation,
		A:         req.A,
		B:         req.B,
		Result:    result,
	}
	if err := db.Create(&calc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Detail: "Failed to create calculation"})
		return
	}

	c.JSON(http.StatusCreated, CalculationResponse{
		ID:        calc.ID,
		Operation: calc.Operation,
		A:         calc.A,
		B:         calc.B,
		Result:    calc.Result,
		CreatedAt: calc.CreatedAt,
	})
}

// listCalculations handles GET /calculations.
func listCalculations(c *gin.Context) {
	var calculations []Calculation
	if err := db.Order("created_at desc").Find(&calculations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Detail: "Failed to list calculations"})
		return
	}

	response := make([]CalculationResponse, len(calculations))
	for i, calc := range calculations {
		response[i] = CalculationResponse{
			ID:        calc.ID,
			Operation: calc.Operation,
			A:         calc.A,
			B:         calc.B,
			Result:    calc.Result,
			CreatedAt: calc.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// getCalculation handles GET /calculations/:id.
func getCalculation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Detail: "Invalid calculation ID"})
		return
	}

	var calc Calculation
	if err := db.First(&calc, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Detail: "Calculation not found"})
		return
	}

	c.JSON(http.StatusOK, CalculationResponse{
		ID:        calc.ID,
		Operation: calc.Operation,
		A:         calc.A,
		B:         calc.B,
		Result:    calc.Result,
		CreatedAt: calc.CreatedAt,
	})
}

// deleteCalculation handles DELETE /calculations/:id.
func deleteCalculation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Detail: "Invalid calculation ID"})
		return
	}

	var calc Calculation
	if err := db.First(&calc, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Detail: "Calculation not found"})
		return
	}

	if err := db.Delete(&calc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Detail: "Failed to delete calculation"})
		return
	}

	c.Status(http.StatusNoContent)
}
