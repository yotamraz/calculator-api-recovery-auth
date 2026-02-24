package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// --- Health Check ---

// healthCheck handles GET /health
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: "0.1.0",
	})
}

// --- Auth Endpoints ---

// register handles POST /auth/register
func register(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserCreate
		if err := c.ShouldBindJSON(&req); err != nil {
			jsonError(c, http.StatusUnprocessableEntity, "Invalid request body")
			return
		}

		// Check for duplicate username
		var existing User
		if err := db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
			jsonError(c, http.StatusBadRequest, "Username already taken")
			return
		}

		// Hash password
		hashed, err := hashPassword(req.Password)
		if err != nil {
			jsonError(c, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		// Create user
		user := User{
			Username:       req.Username,
			HashedPassword: hashed,
			CreatedAt:      time.Now().UTC(),
		}
		if err := db.Create(&user).Error; err != nil {
			jsonError(c, http.StatusInternalServerError, "Failed to create user")
			return
		}

		c.JSON(http.StatusCreated, UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		})
	}
}

// login handles POST /auth/token (form-encoded)
func login(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form LoginForm
		if err := c.ShouldBind(&form); err != nil {
			jsonError(c, http.StatusUnprocessableEntity, "Invalid request body")
			return
		}

		// Authenticate user
		var user User
		if err := db.Where("username = ?", form.Username).First(&user).Error; err != nil {
			jsonError(c, http.StatusUnauthorized, "Incorrect username or password")
			return
		}

		if !verifyPassword(form.Password, user.HashedPassword) {
			jsonError(c, http.StatusUnauthorized, "Incorrect username or password")
			return
		}

		// Create token
		token, err := createAccessToken(user.Username, 0)
		if err != nil {
			jsonError(c, http.StatusInternalServerError, "Failed to create token")
			return
		}

		c.JSON(http.StatusOK, Token{
			AccessToken: token,
			TokenType:   "bearer",
		})
	}
}

// --- Calculator Endpoints ---

// apiAdd handles POST /add
func apiAdd(c *gin.Context) {
	var req CalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusUnprocessableEntity, "Invalid request body")
		return
	}
	c.JSON(http.StatusOK, ResultResponse{Result: add(*req.A, *req.B)})
}

// apiSubtract handles POST /subtract
func apiSubtract(c *gin.Context) {
	var req CalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusUnprocessableEntity, "Invalid request body")
		return
	}
	c.JSON(http.StatusOK, ResultResponse{Result: subtract(*req.A, *req.B)})
}

// apiMultiply handles POST /multiply
func apiMultiply(c *gin.Context) {
	var req CalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusUnprocessableEntity, "Invalid request body")
		return
	}
	c.JSON(http.StatusOK, ResultResponse{Result: multiply(*req.A, *req.B)})
}

// apiDivide handles POST /divide
func apiDivide(c *gin.Context) {
	var req CalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusUnprocessableEntity, "Invalid request body")
		return
	}
	result, err := divide(*req.A, *req.B)
	if err != nil {
		jsonError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, ResultResponse{Result: result})
}

// --- Calculations CRUD Endpoints ---

// Operations maps operation names to their functions.
var operations = map[string]func(float64, float64) float64{
	"add": add,
	"sub": subtract,
	"mul": multiply,
}

// createCalculation handles POST /calculations
func createCalculation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CalculationCreate
		if err := c.ShouldBindJSON(&req); err != nil {
			jsonError(c, http.StatusUnprocessableEntity, "Invalid request body")
			return
		}

		var result float64

		// Handle division separately due to error return
		if req.Operation == "div" {
			var err error
			result, err = divide(*req.A, *req.B)
			if err != nil {
				jsonError(c, http.StatusBadRequest, err.Error())
				return
			}
		} else if opFunc, ok := operations[req.Operation]; ok {
			result = opFunc(*req.A, *req.B)
		} else {
			jsonError(c, http.StatusBadRequest, fmt.Sprintf("Unknown operation: %s. Use: [add sub mul div]", req.Operation))
			return
		}

		calc := Calculation{
			Operation: req.Operation,
			A:         *req.A,
			B:         *req.B,
			Result:    result,
		}
		if err := db.Create(&calc).Error; err != nil {
			jsonError(c, http.StatusInternalServerError, "Failed to create calculation")
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
}

// listCalculations handles GET /calculations
func listCalculations(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var calculations []Calculation
		if err := db.Order("created_at desc").Find(&calculations).Error; err != nil {
			jsonError(c, http.StatusInternalServerError, "Failed to fetch calculations")
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
}

// getCalculation handles GET /calculations/:id
func getCalculation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var calc Calculation
		if err := db.First(&calc, id).Error; err != nil {
			jsonError(c, http.StatusNotFound, "Calculation not found")
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
}

// deleteCalculation handles DELETE /calculations/:id
func deleteCalculation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var calc Calculation
		if err := db.First(&calc, id).Error; err != nil {
			jsonError(c, http.StatusNotFound, "Calculation not found")
			return
		}

		if err := db.Delete(&calc).Error; err != nil {
			jsonError(c, http.StatusInternalServerError, "Failed to delete calculation")
			return
		}

		c.Status(http.StatusNoContent)
	}
}
