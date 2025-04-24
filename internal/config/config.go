package config

import (
	"encoding/json"
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
	Port     string `json:"port"`
	DBUrl    string `json:"db_url"`
	LogLevel string `json:"log_level"` // e.g. "info", "debug", "warn", "error"
}

type AppConfig struct {
	Google GoogleConfig `json:"google"`
	OpenAI OpenAIConfig `json:"openai"`
	Server ServerConfig `json:"server"`
}

func LoadConfig(path string) (*AppConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var cfg AppConfig
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
