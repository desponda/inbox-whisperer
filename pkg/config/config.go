package config

import (
	"github.com/rs/zerolog/log"
	"os"
)

// Config holds all configuration for the server
type Config struct {
	Port     string
	DBUrl    string
	LogLevel string
}

// Load configuration from environment variables with sensible defaults.
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

// Print all config values for debugging
func (c *Config) Print() {
	log.Info().Str("port", c.Port).Str("database_url", c.DBUrl).Str("log_level", c.LogLevel).Msg("Config loaded")
}
