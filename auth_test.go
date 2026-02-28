package main

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key"

// ============================================================================
// Password Hashing Tests
// ============================================================================

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == password {
		t.Fatal("HashPassword returned the original password, not a hash")
	}
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	password := "samepassword"
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)
	if hash1 == hash2 {
		t.Error("HashPassword returned the same hash for the same password (salt should differ)")
	}
}

func TestVerifyPasswordCorrect(t *testing.T) {
	password := "correctpassword"
	hash, _ := HashPassword(password)
	if !VerifyPassword(password, hash) {
		t.Error("VerifyPassword returned false for correct password")
	}
}

func TestVerifyPasswordIncorrect(t *testing.T) {
	password := "correctpassword"
	hash, _ := HashPassword(password)
	if VerifyPassword("wrongpassword", hash) {
		t.Error("VerifyPassword returned true for incorrect password")
	}
}

func TestVerifyPasswordEmptyPassword(t *testing.T) {
	hash, _ := HashPassword("somepassword")
	if VerifyPassword("", hash) {
		t.Error("VerifyPassword returned true for empty password")
	}
}

// ============================================================================
// JWT Token Tests
// ============================================================================

func TestCreateAccessToken(t *testing.T) {
	tokenString, err := CreateAccessToken("testuser", testSecret, 30)
	if err != nil {
		t.Fatalf("CreateAccessToken returned error: %v", err)
	}
	if tokenString == "" {
		t.Fatal("CreateAccessToken returned empty token")
	}
}

func TestCreateAccessTokenAndParse(t *testing.T) {
	username := "testuser"
	tokenString, err := CreateAccessToken(username, testSecret, 30)
	if err != nil {
		t.Fatalf("CreateAccessToken returned error: %v", err)
	}

	parsedUsername, err := ParseToken(tokenString, testSecret)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if parsedUsername != username {
		t.Errorf("ParseToken returned username %q, want %q", parsedUsername, username)
	}
}

func TestParseTokenInvalid(t *testing.T) {
	_, err := ParseToken("invalid-token-string", testSecret)
	if err == nil {
		t.Error("ParseToken should return error for invalid token")
	}
}

func TestParseTokenWrongSecret(t *testing.T) {
	tokenString, _ := CreateAccessToken("testuser", testSecret, 30)
	_, err := ParseToken(tokenString, "wrong-secret")
	if err == nil {
		t.Error("ParseToken should return error for wrong secret")
	}
}

func TestParseTokenExpired(t *testing.T) {
	// Create a token that expired 1 hour ago
	claims := jwt.MapClaims{
		"sub": "testuser",
		"exp": jwt.NewNumericDate(time.Now().UTC().Add(-1 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = ParseToken(tokenString, testSecret)
	if err == nil {
		t.Error("ParseToken should return error for expired token")
	}
}

func TestParseTokenMissingSub(t *testing.T) {
	// Create a token without a "sub" claim
	claims := jwt.MapClaims{
		"exp": jwt.NewNumericDate(time.Now().UTC().Add(30 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testSecret))

	_, err := ParseToken(tokenString, testSecret)
	if err == nil {
		t.Error("ParseToken should return error for missing sub claim")
	}
}

func TestParseTokenEmptySub(t *testing.T) {
	// Create a token with an empty "sub" claim
	claims := jwt.MapClaims{
		"sub": "",
		"exp": jwt.NewNumericDate(time.Now().UTC().Add(30 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testSecret))

	_, err := ParseToken(tokenString, testSecret)
	if err == nil {
		t.Error("ParseToken should return error for empty sub claim")
	}
}

// ============================================================================
// AuthenticateUser Tests
// ============================================================================

func TestAuthenticateUserValid(t *testing.T) {
	db := InitDB(":memory:")

	hash, _ := HashPassword("password123")
	db.Create(&User{Username: "testuser", HashedPassword: hash})

	user := AuthenticateUser(db, "testuser", "password123")
	if user == nil {
		t.Fatal("AuthenticateUser returned nil for valid credentials")
	}
	if user.Username != "testuser" {
		t.Errorf("AuthenticateUser returned username %q, want %q", user.Username, "testuser")
	}
}

func TestAuthenticateUserWrongPassword(t *testing.T) {
	db := InitDB(":memory:")

	hash, _ := HashPassword("password123")
	db.Create(&User{Username: "testuser", HashedPassword: hash})

	user := AuthenticateUser(db, "testuser", "wrongpassword")
	if user != nil {
		t.Error("AuthenticateUser should return nil for wrong password")
	}
}

func TestAuthenticateUserNotFound(t *testing.T) {
	db := InitDB(":memory:")

	user := AuthenticateUser(db, "nonexistent", "password123")
	if user != nil {
		t.Error("AuthenticateUser should return nil for non-existent user")
	}
}
