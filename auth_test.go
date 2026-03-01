package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// --- Password Hashing Tests ---

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("mysecretpassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Error("HashPassword returned empty string")
	}
	// bcrypt hashes start with "$2a$" or "$2b$"
	if hash[0] != '$' {
		t.Errorf("HashPassword returned non-bcrypt hash: %s", hash)
	}
}

func TestHashPasswordDifferentSalts(t *testing.T) {
	hash1, _ := HashPassword("samepassword")
	hash2, _ := HashPassword("samepassword")
	if hash1 == hash2 {
		t.Error("Two hashes of the same password should be different (different salts)")
	}
}

func TestVerifyPasswordCorrect(t *testing.T) {
	password := "correctpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if !VerifyPassword(password, hash) {
		t.Error("VerifyPassword should return true for correct password")
	}
}

func TestVerifyPasswordWrong(t *testing.T) {
	hash, err := HashPassword("correctpassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if VerifyPassword("wrongpassword", hash) {
		t.Error("VerifyPassword should return false for wrong password")
	}
}

func TestVerifyPasswordEmpty(t *testing.T) {
	hash, err := HashPassword("somepassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if VerifyPassword("", hash) {
		t.Error("VerifyPassword should return false for empty password")
	}
}

// --- JWT Token Tests ---

func TestCreateAccessToken(t *testing.T) {
	cfg := Config{
		JWTSecretKey:         "test-secret-key",
		AccessTokenExpireMin: 30,
	}

	tokenString, err := CreateAccessToken("alice", cfg)
	if err != nil {
		t.Fatalf("CreateAccessToken returned error: %v", err)
	}
	if tokenString == "" {
		t.Error("CreateAccessToken returned empty token")
	}

	// Parse and verify the token
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecretKey), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}
	if !token.Valid {
		t.Error("Token should be valid")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		t.Fatal("Failed to cast claims")
	}
	if claims.Subject != "alice" {
		t.Errorf("Token subject = %q, want %q", claims.Subject, "alice")
	}
}

func TestCreateAccessTokenExpiry(t *testing.T) {
	cfg := Config{
		JWTSecretKey:         "test-secret-key",
		AccessTokenExpireMin: 60,
	}

	before := time.Now().UTC().Truncate(time.Second)
	tokenString, err := CreateAccessToken("bob", cfg)
	if err != nil {
		t.Fatalf("CreateAccessToken returned error: %v", err)
	}
	after := time.Now().UTC().Add(time.Second).Truncate(time.Second)

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecretKey), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	claims := token.Claims.(*jwt.RegisteredClaims)
	expiry := claims.ExpiresAt.Time

	// JWT exp is stored as Unix timestamp (integer seconds), so we compare
	// with second-level precision using truncated boundaries.
	expectedEarliest := before.Add(60 * time.Minute)
	expectedLatest := after.Add(60 * time.Minute)

	if expiry.Before(expectedEarliest) || expiry.After(expectedLatest) {
		t.Errorf("Token expiry = %v, expected between %v and %v", expiry, expectedEarliest, expectedLatest)
	}
}

func TestCreateAccessTokenSigningMethod(t *testing.T) {
	cfg := Config{
		JWTSecretKey:         "test-secret-key",
		AccessTokenExpireMin: 30,
	}

	tokenString, _ := CreateAccessToken("alice", cfg)

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecretKey), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if token.Method.Alg() != "HS256" {
		t.Errorf("Token signing method = %q, want %q", token.Method.Alg(), "HS256")
	}
}

// --- AuthMiddleware Tests ---

// setupMiddlewareTestRouter creates a simple router with a protected test endpoint.
func setupMiddlewareTestRouter() (*gin.Engine, Config) {
	gin.SetMode(gin.TestMode)
	cfg := Config{
		JWTSecretKey:         "test-secret-key",
		DatabaseURL:          ":memory:",
		AccessTokenExpireMin: 30,
	}
	db := InitDB(cfg.DatabaseURL)

	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg))
	protected.GET("/protected", func(c *gin.Context) {
		user := GetCurrentUser(c)
		c.JSON(http.StatusOK, gin.H{"username": user.Username})
	})

	return r, cfg
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	router, _ := setupMiddlewareTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Not authenticated" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Not authenticated")
	}

	wwwAuth := w.Header().Get("WWW-Authenticate")
	if wwwAuth != "Bearer" {
		t.Errorf("WWW-Authenticate = %q, want %q", wwwAuth, "Bearer")
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router, _ := setupMiddlewareTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Could not validate credentials" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Could not validate credentials")
	}
}

func TestAuthMiddleware_MalformedAuthHeader(t *testing.T) {
	router, _ := setupMiddlewareTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := Config{
		JWTSecretKey:         "test-secret-key",
		DatabaseURL:          ":memory:",
		AccessTokenExpireMin: 30,
	}
	db := InitDB(cfg.DatabaseURL)

	// Create a user
	hash, _ := HashPassword("pass")
	db.Create(&User{Username: "alice", HashedPassword: hash})

	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg))
	protected.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Create an expired token manually
	claims := jwt.RegisteredClaims{
		Subject:   "alice",
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(-1 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, _ := token.SignedString([]byte(cfg.JWTSecretKey))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Could not validate credentials" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Could not validate credentials")
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := Config{
		JWTSecretKey:         "test-secret-key",
		DatabaseURL:          ":memory:",
		AccessTokenExpireMin: 30,
	}
	db := InitDB(cfg.DatabaseURL)

	// Create a user in the database
	hash, _ := HashPassword("password123")
	db.Create(&User{Username: "alice", HashedPassword: hash})

	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg))
	protected.GET("/protected", func(c *gin.Context) {
		user := GetCurrentUser(c)
		c.JSON(http.StatusOK, gin.H{"username": user.Username})
	})

	// Create a valid token
	tokenString, _ := CreateAccessToken("alice", cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["username"] != "alice" {
		t.Errorf("Username = %q, want %q", resp["username"], "alice")
	}
}

func TestAuthMiddleware_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := Config{
		JWTSecretKey:         "test-secret-key",
		DatabaseURL:          ":memory:",
		AccessTokenExpireMin: 30,
	}
	db := InitDB(cfg.DatabaseURL)

	// Do NOT create a user — token references a non-existent user

	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg))
	protected.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Create a valid token for a user that doesn't exist in DB
	tokenString, _ := CreateAccessToken("nonexistent", cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["detail"] != "Could not validate credentials" {
		t.Errorf("Error detail = %q, want %q", resp["detail"], "Could not validate credentials")
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	router, _ := setupMiddlewareTestRouter()

	// Create a token with a different secret
	wrongCfg := Config{
		JWTSecretKey:         "wrong-secret",
		AccessTokenExpireMin: 30,
	}
	tokenString, _ := CreateAccessToken("alice", wrongCfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_EmptySubject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := Config{
		JWTSecretKey:         "test-secret-key",
		DatabaseURL:          ":memory:",
		AccessTokenExpireMin: 30,
	}
	db := InitDB(cfg.DatabaseURL)

	r := gin.New()
	protected := r.Group("/")
	protected.Use(AuthMiddleware(db, cfg))
	protected.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Create a token with empty subject
	claims := jwt.RegisteredClaims{
		Subject:   "",
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(30 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecretKey))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// --- AuthenticateUser Tests ---

func TestAuthenticateUser_Valid(t *testing.T) {
	db := InitDB(":memory:")
	hash, _ := HashPassword("correctpass")
	db.Create(&User{Username: "testuser", HashedPassword: hash})

	user := AuthenticateUser(db, "testuser", "correctpass")
	if user == nil {
		t.Fatal("AuthenticateUser should return a user for valid credentials")
	}
	if user.Username != "testuser" {
		t.Errorf("Username = %q, want %q", user.Username, "testuser")
	}
}

func TestAuthenticateUser_WrongPassword(t *testing.T) {
	db := InitDB(":memory:")
	hash, _ := HashPassword("correctpass")
	db.Create(&User{Username: "testuser", HashedPassword: hash})

	user := AuthenticateUser(db, "testuser", "wrongpass")
	if user != nil {
		t.Error("AuthenticateUser should return nil for wrong password")
	}
}

func TestAuthenticateUser_NonexistentUser(t *testing.T) {
	db := InitDB(":memory:")

	user := AuthenticateUser(db, "nobody", "somepass")
	if user != nil {
		t.Error("AuthenticateUser should return nil for non-existent user")
	}
}

// --- GetCurrentUser/SetCurrentUser Tests ---

func TestSetAndGetCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	user := &User{ID: 1, Username: "alice"}
	SetCurrentUser(c, user)

	got := GetCurrentUser(c)
	if got == nil {
		t.Fatal("GetCurrentUser returned nil")
	}
	if got.Username != "alice" {
		t.Errorf("Username = %q, want %q", got.Username, "alice")
	}
}

func TestGetCurrentUser_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	got := GetCurrentUser(c)
	if got != nil {
		t.Error("GetCurrentUser should return nil when no user is set")
	}
}
