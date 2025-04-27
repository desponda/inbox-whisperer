package config

import (
	"os"
	"testing"
)

func TestLoadConfig_MissingFile(t *testing.T) {
	cfg, err := LoadConfig("/tmp/definitely-does-not-exist.json")
	if err != nil {
		t.Errorf("did not expect error for missing config file, got: %v", err)
	}
	if cfg == nil {
		t.Error("expected non-nil config when falling back to env vars")
	}
}

func TestLoadConfig_MalformedJSON(t *testing.T) {
	f, err := os.CreateTemp("", "bad_config_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("{not valid json}"); err != nil {
		t.Fatalf("failed to write invalid json: %v", err)
	}
	f.Close()

	_, err = LoadConfig(f.Name())
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func TestLoadConfig_ValidConfig(t *testing.T) {
	cfgText := `{
	  "google": {
	    "client_id": "id",
	    "client_secret": "secret",
	    "redirect_url": "http://localhost"
	  },
	  "openai": {
	    "api_key": "sk-abc"
	  },
	  "server": {
	    "port": "8080",
	    "db_url": "postgres://user:pass@localhost/db",
	    "log_level": "debug"
	  }
	}`
	f, err := os.CreateTemp("", "good_config_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(cfgText); err != nil {
		t.Fatalf("failed to write config text: %v", err)
	}
	f.Close()

	cfg, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("unexpected error loading valid config: %v", err)
	}
	if cfg.Google.ClientID != "id" {
		t.Errorf("expected google.client_id 'id', got '%s'", cfg.Google.ClientID)
	}
	if cfg.Google.ClientSecret != "secret" {
		t.Errorf("expected google.client_secret 'secret', got '%s'", cfg.Google.ClientSecret)
	}
	if cfg.Google.RedirectURL != "http://localhost" {
		t.Errorf("expected google.redirect_url 'http://localhost', got '%s'", cfg.Google.RedirectURL)
	}
	if cfg.OpenAI.APIKey != "sk-abc" {
		t.Errorf("expected openai.api_key 'sk-abc', got '%s'", cfg.OpenAI.APIKey)
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("expected server.port '8080', got '%s'", cfg.Server.Port)
	}
	if cfg.Server.DBUrl != "postgres://user:pass@localhost/db" {
		t.Errorf("expected server.db_url 'postgres://user:pass@localhost/db', got '%s'", cfg.Server.DBUrl)
	}
	if cfg.Server.LogLevel != "debug" {
		t.Errorf("expected server.log_level 'debug', got '%s'", cfg.Server.LogLevel)
	}
}
