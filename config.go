package main

import (
	"os"
	"strconv"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	JWTSecretKey         string
	DatabaseURL          string
	AccessTokenExpireMin int
}

// LoadConfig reads configuration from environment variables with sensible defaults
// matching the original Python application.
func LoadConfig() Config {
	cfg := Config{
		JWTSecretKey:         "dev-secret-key-change-me-in-production",
		DatabaseURL:          "file::memory:?cache=shared",
		AccessTokenExpireMin: 30,
	}

	if v := os.Getenv("JWT_SECRET_KEY"); v != "" {
		cfg.JWTSecretKey = v
	}

	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.DatabaseURL = v
	}

	if v := os.Getenv("ACCESS_TOKEN_EXPIRE_MINUTES"); v != "" {
		if minutes, err := strconv.Atoi(v); err == nil && minutes > 0 {
			cfg.AccessTokenExpireMin = minutes
		}
	}

	return cfg
}
