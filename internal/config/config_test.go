package config

import (
	"os"
	"testing"
)

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/tmp/definitely-does-not-exist.json")
	if err == nil {
		t.Error("expected error for missing config file, got nil")
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
	cfgText := `{"google":{"client_id":"id","client_secret":"secret","redirect_url":"http://localhost"},"openai":{"api_key":"sk-abc"},"server":{"port":"8080","db_url":"postgres://user:pass@localhost/db"}}`
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
	if cfg.Server.Port != "8080" || cfg.Google.ClientID != "id" {
		t.Errorf("unexpected config values: %+v", cfg)
	}
}
