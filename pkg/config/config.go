package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the server
// Add new fields as needed for DB, secrets, etc.
type Config struct {
	Port     string
	DBUrl    string
	LogLevel string
}

// Load loads configuration from environment variables, with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Port:     getEnv("PORT", "8080"),
		DBUrl:    getEnv("DATABASE_URL", "postgres://inbox:inbox@localhost:5432/inboxwhisperer?sslmode=disable"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
	return cfg, nil
}

// getEnv returns the value of the environment variable or a default.
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Print prints all config values for debugging
func (c *Config) Print() {
	fmt.Printf("Config: PORT=%s, DATABASE_URL=%s, LOG_LEVEL=%s\n", c.Port, c.DBUrl, c.LogLevel)
}
