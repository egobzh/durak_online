package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort       string
	DatabaseURL      string
	JWTSecret        string
	JWTExpiration    time.Duration
	ResetBaseURL     string
	ResetTokenTTLMin time.Duration
}

func LoadConfig() *Config {
	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/durak_db?sslmode=disable"),
		JWTSecret:        getEnv("JWT_SECRET", "super-secret-durak-jwt-key-2026"),
		JWTExpiration:    getEnvAsDuration("JWT_EXPIRATION_HOURS", 24) * time.Hour,
		ResetBaseURL:     getEnv("RESET_BASE_URL", "http://localhost:8080/api/v1/auth/reset-password"),
		ResetTokenTTLMin: getEnvAsDuration("RESET_TOKEN_TTL_MINUTES", 30) * time.Minute,
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal int) time.Duration {
	if valStr, ok := os.LookupEnv(key); ok {
		if val, err := strconv.Atoi(valStr); err == nil {
			return time.Duration(val)
		}
	}
	return time.Duration(defaultVal)
}
