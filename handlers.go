package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(v) //nolint:errcheck
}

// writeError writes a JSON error response matching FastAPI's {"detail": "..."} format.
func writeError(w http.ResponseWriter, status int, detail string) {
	writeJSON(w, status, ErrorResponse{Detail: detail})
}

// writeValidationError writes a 422 JSON error response matching FastAPI/Pydantic's
// validation error format where "detail" is an array of error items.
func writeValidationError(w http.ResponseWriter, fields []ValidationErrorItem) {
	writeJSON(w, http.StatusUnprocessableEntity, ValidationErrorResponse{Detail: fields})
}

// HealthHandler handles GET /health.
func (d *Deps) HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: "0.1.0",
	})
}

// ---------------------------------------------------------------------------
// Auth Handlers
// ---------------------------------------------------------------------------

// RegisterHandler handles POST /auth/register.
// Accepts JSON {"username": "...", "password": "..."}, creates a new user with
// a bcrypt-hashed password, and returns 201 with the UserResponse.
// Returns 400 if the username is already taken.
// Equivalent to Python server.py register().
func (d *Deps) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Decode into both the typed struct and a raw map so we can echo the input
	// in validation errors (matching Pydantic v2 behavior).
	var raw map[string]any
	var req UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeValidationError(w, []ValidationErrorItem{
			{Loc: []string{"body"}, Msg: "Invalid request body", Type: "value_error", Input: nil},
		})
		return
	}

	// Marshal/unmarshal from raw map into the typed struct.
	req.Username, _ = raw["username"].(string)
	req.Password, _ = raw["password"].(string)

	// Build a sanitized copy of the input for validation errors, masking
	// sensitive fields like "password" (matching Pydantic v2 / FastAPI behavior
	// where SecretStr or test-framework sanitization masks passwords).
	sanitizedInput := make(map[string]any, len(raw))
	for k, v := range raw {
		if k == "password" {
			sanitizedInput[k] = "********"
		} else {
			sanitizedInput[k] = v
		}
	}

	// Validate required fields (mirroring Pydantic v2's required-field validation).
	var validationErrors []ValidationErrorItem
	if _, ok := raw["username"]; !ok || req.Username == "" {
		validationErrors = append(validationErrors, ValidationErrorItem{
			Loc:   []string{"body", "username"},
			Msg:   "Field required",
			Type:  "missing",
			Input: sanitizedInput,
		})
	}
	if _, ok := raw["password"]; !ok || req.Password == "" {
		validationErrors = append(validationErrors, ValidationErrorItem{
			Loc:   []string{"body", "password"},
			Msg:   "Field required",
			Type:  "missing",
			Input: sanitizedInput,
		})
	}
	if len(validationErrors) > 0 {
		writeValidationError(w, validationErrors)
		return
	}

	// Check for duplicate username.
	var existing User
	if err := d.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		writeError(w, http.StatusBadRequest, "Username already taken")
		return
	}

	// Hash the password.
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create the user.
	user := User{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		CreatedAt:      time.Now().UTC(),
	}
	if err := d.DB.Create(&user).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Return the created user (without password).
	writeJSON(w, http.StatusCreated, UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	})
}

// TokenHandler handles POST /auth/token.
// Accepts form-encoded data (username + password) matching OAuth2PasswordRequestForm.
// Returns {"access_token": "...", "token_type": "bearer"} on success.
// Returns 401 with "Incorrect username or password" on failure.
// Equivalent to Python server.py login().
func (d *Deps) TokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data (matching OAuth2PasswordRequestForm behavior).
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Authenticate the user.
	user := AuthenticateUser(d.DB, username, password)
	if user == nil {
		writeAuthError(w, "Incorrect username or password")
		return
	}

	// Create the access token.
	accessToken, err := CreateAccessToken(user.Username, d.Config.JWTSecretKey, AccessTokenExpireMinutes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create access token")
		return
	}

	writeJSON(w, http.StatusOK, Token{
		AccessToken: accessToken,
		TokenType:   "bearer",
	})
}
