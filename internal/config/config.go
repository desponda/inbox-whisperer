package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type GoogleConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

type OpenAIConfig struct {
	APIKey string `json:"api_key"`
}

type ServerConfig struct {
	Port                string `json:"port"`
	DBUrl               string `json:"db_url"`
	LogLevel            string `json:"log_level"` // e.g. "info", "debug", "warn", "error"
	FrontendURL         string `json:"frontend_url"`
	SessionKey          string `json:"session_key"` // Key for signing session cookies
	SessionCookieSecure bool   `json:"sessionCookieSecure"`
}

type AppConfig struct {
	Google GoogleConfig `json:"google"`
	OpenAI OpenAIConfig `json:"openai"`
	Server ServerConfig `json:"server"`
}

func LoadConfig(path string) (*AppConfig, error) {
	// Log the config path
	fmt.Fprintf(os.Stderr, "[DEBUG] Attempting to load config from: %s\n", path)

	// Try to stat the file
	info, err := os.Stat(path)
	if err == nil && !info.IsDir() {
		// File exists, try to load it
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to open %s: %v\n", path, err)
			return nil, err
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		var cfg AppConfig
		if err := dec.Decode(&cfg); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to decode config JSON: %v\n", err)
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] Config loaded successfully from %s\n", path)
		return &cfg, nil
	}

	fmt.Fprintf(os.Stderr, "[WARN] Config file not found at %s, attempting to load from environment variables\n", path)
	cfg := AppConfig{
		Google: GoogleConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		},
		OpenAI: OpenAIConfig{
			APIKey: os.Getenv("OPENAI_API_KEY"),
		},
		Server: ServerConfig{
			Port:                os.Getenv("SERVER_PORT"),
			DBUrl:               os.Getenv("DATABASE_URL"),
			LogLevel:            os.Getenv("LOG_LEVEL"),
			SessionKey:          os.Getenv("SESSION_KEY"),
			SessionCookieSecure: os.Getenv("SESSION_COOKIE_SECURE") == "true",
		},
	}
	return &cfg, nil
}
