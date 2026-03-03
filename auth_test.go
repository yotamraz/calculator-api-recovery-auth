package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ---------------------------------------------------------------------------
// Password hashing & verification tests
// ---------------------------------------------------------------------------

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" {
		t.Error("HashPassword() returned empty string")
	}
	// Hash should be different from the plain password.
	if hash == "testpassword123" {
		t.Error("HashPassword() returned the plain password")
	}
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	// Two hashes of the same password should be different (salt).
	hash1, _ := HashPassword("samepassword")
	hash2, _ := HashPassword("samepassword")
	if hash1 == hash2 {
		t.Error("HashPassword() produced identical hashes for same password (salt not working)")
	}
}

func TestVerifyPasswordCorrect(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if !VerifyPassword("mypassword", hash) {
		t.Error("VerifyPassword() returned false for correct password")
	}
}

func TestVerifyPasswordIncorrect(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if VerifyPassword("wrongpassword", hash) {
		t.Error("VerifyPassword() returned true for incorrect password")
	}
}

func TestVerifyPasswordEmptyPassword(t *testing.T) {
	hash, err := HashPassword("notempty")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if VerifyPassword("", hash) {
		t.Error("VerifyPassword() returned true for empty password")
	}
}

// ---------------------------------------------------------------------------
// JWT creation & parsing tests
// ---------------------------------------------------------------------------

func TestCreateAccessToken(t *testing.T) {
	token, err := CreateAccessToken("testuser", "test-secret", 30)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}
	if token == "" {
		t.Error("CreateAccessToken() returned empty string")
	}
}

func TestParseTokenValid(t *testing.T) {
	secret := "test-secret-key"
	token, err := CreateAccessToken("testuser", secret, 30)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	username, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if username != "testuser" {
		t.Errorf("ParseToken() username = %q, want %q", username, "testuser")
	}
}

func TestParseTokenExpired(t *testing.T) {
	secret := "test-secret-key"

	// Create a token that expired 1 minute ago.
	claims := jwt.MapClaims{
		"sub": "testuser",
		"exp": jwt.NewNumericDate(time.Now().UTC().Add(-1 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}

	_, err = ParseToken(tokenString, secret)
	if err == nil {
		t.Error("ParseToken() expected error for expired token, got nil")
	}
}

func TestParseTokenInvalidSignature(t *testing.T) {
	token, err := CreateAccessToken("testuser", "correct-secret", 30)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	_, err = ParseToken(token, "wrong-secret")
	if err == nil {
		t.Error("ParseToken() expected error for wrong secret, got nil")
	}
}

func TestParseTokenMissingSub(t *testing.T) {
	secret := "test-secret-key"

	// Create a token without "sub" claim.
	claims := jwt.MapClaims{
		"exp": jwt.NewNumericDate(time.Now().UTC().Add(30 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	_, err = ParseToken(tokenString, secret)
	if err == nil {
		t.Error("ParseToken() expected error for missing sub claim, got nil")
	}
}

func TestParseTokenEmptySub(t *testing.T) {
	secret := "test-secret-key"

	claims := jwt.MapClaims{
		"sub": "",
		"exp": jwt.NewNumericDate(time.Now().UTC().Add(30 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	_, err = ParseToken(tokenString, secret)
	if err == nil {
		t.Error("ParseToken() expected error for empty sub claim, got nil")
	}
}

func TestParseTokenInvalidString(t *testing.T) {
	_, err := ParseToken("not-a-valid-jwt", "test-secret")
	if err == nil {
		t.Error("ParseToken() expected error for invalid token string, got nil")
	}
}

// ---------------------------------------------------------------------------
// AuthenticateUser tests
// ---------------------------------------------------------------------------

func TestAuthenticateUserValid(t *testing.T) {
	deps := newTestDeps(t)

	// Create a test user.
	hash, _ := HashPassword("password123")
	user := User{Username: "authuser", HashedPassword: hash}
	deps.DB.Create(&user)

	result := AuthenticateUser(deps.DB, "authuser", "password123")
	if result == nil {
		t.Fatal("AuthenticateUser() returned nil for valid credentials")
	}
	if result.Username != "authuser" {
		t.Errorf("AuthenticateUser() username = %q, want %q", result.Username, "authuser")
	}
}

func TestAuthenticateUserInvalidPassword(t *testing.T) {
	deps := newTestDeps(t)

	hash, _ := HashPassword("correctpassword")
	user := User{Username: "authuser2", HashedPassword: hash}
	deps.DB.Create(&user)

	result := AuthenticateUser(deps.DB, "authuser2", "wrongpassword")
	if result != nil {
		t.Error("AuthenticateUser() expected nil for wrong password")
	}
}

func TestAuthenticateUserNonexistentUser(t *testing.T) {
	deps := newTestDeps(t)

	result := AuthenticateUser(deps.DB, "nonexistent", "password")
	if result != nil {
		t.Error("AuthenticateUser() expected nil for nonexistent user")
	}
}

// ---------------------------------------------------------------------------
// UserFromContext tests
// ---------------------------------------------------------------------------

func TestUserFromContextPresent(t *testing.T) {
	deps := newTestDeps(t)

	// Create a test user in the DB.
	hash, _ := HashPassword("pass")
	testUser := User{Username: "ctxuser", HashedPassword: hash}
	deps.DB.Create(&testUser)

	// Create a valid token.
	token, _ := CreateAccessToken("ctxuser", deps.Config.JWTSecretKey, 30)

	// Set up a protected handler that reads from context.
	handler := deps.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())
		if u == nil {
			t.Error("UserFromContext() returned nil inside auth middleware")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if u.Username != "ctxuser" {
			t.Errorf("UserFromContext() username = %q, want %q", u.Username, "ctxuser")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handler status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestUserFromContextAbsent(t *testing.T) {
	// Without middleware, context should not have a user.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	u := UserFromContext(req.Context())
	if u != nil {
		t.Error("UserFromContext() expected nil when no user in context")
	}
}

// ---------------------------------------------------------------------------
// AuthMiddleware tests
// ---------------------------------------------------------------------------

func TestAuthMiddlewareNoToken(t *testing.T) {
	deps := newTestDeps(t)

	handler := deps.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("middleware status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	// Check WWW-Authenticate header.
	if wwwAuth := w.Header().Get("WWW-Authenticate"); wwwAuth != "Bearer" {
		t.Errorf("WWW-Authenticate = %q, want %q", wwwAuth, "Bearer")
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	deps := newTestDeps(t)

	handler := deps.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-jwt-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("middleware status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareExpiredToken(t *testing.T) {
	deps := newTestDeps(t)

	// Create a user so DB lookup would succeed if token were valid.
	hash, _ := HashPassword("pass")
	deps.DB.Create(&User{Username: "expireduser", HashedPassword: hash})

	// Create an expired token.
	claims := jwt.MapClaims{
		"sub": "expireduser",
		"exp": jwt.NewNumericDate(time.Now().UTC().Add(-1 * time.Minute)),
	}
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := tk.SignedString([]byte(deps.Config.JWTSecretKey))

	handler := deps.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("middleware status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareWrongSecret(t *testing.T) {
	deps := newTestDeps(t)

	hash, _ := HashPassword("pass")
	deps.DB.Create(&User{Username: "wrongsecretuser", HashedPassword: hash})

	// Create token with a different secret.
	token, _ := CreateAccessToken("wrongsecretuser", "different-secret", 30)

	handler := deps.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("middleware status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareUserNotInDB(t *testing.T) {
	deps := newTestDeps(t)

	// Create a valid token for a user that doesn't exist in the DB.
	token, _ := CreateAccessToken("ghostuser", deps.Config.JWTSecretKey, 30)

	handler := deps.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("middleware status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareMalformedHeader(t *testing.T) {
	deps := newTestDeps(t)

	handler := deps.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Header without "Bearer " prefix.
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token some-token-value")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("middleware status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	deps := newTestDeps(t)

	hash, _ := HashPassword("pass")
	deps.DB.Create(&User{Username: "validuser", HashedPassword: hash})

	token, _ := CreateAccessToken("validuser", deps.Config.JWTSecretKey, 30)

	called := false
	handler := deps.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		u := UserFromContext(r.Context())
		if u == nil {
			t.Error("UserFromContext() returned nil for valid token")
		} else if u.Username != "validuser" {
			t.Errorf("UserFromContext().Username = %q, want %q", u.Username, "validuser")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Error("next handler was not called for valid token")
	}
	if w.Code != http.StatusOK {
		t.Errorf("middleware status = %d, want %d", w.Code, http.StatusOK)
	}
}
