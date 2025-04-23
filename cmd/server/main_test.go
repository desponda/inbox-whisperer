package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/desponda/inbox-whisperer/internal/config"
)

func TestHealthz(t *testing.T) {
	dummyCfg := &config.AppConfig{
		Google: config.GoogleConfig{
			ClientID:     "dummy",
			ClientSecret: "dummy",
			RedirectURL:  "http://localhost:8080/api/auth/callback",
		},
	}
	r := setupRouter(nil, dummyCfg)
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("could not send GET /healthz: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %v", resp.Status)
	}}

func TestServerStartupWithValidConfig(t *testing.T) {
	cfgText := `{"google":{"client_id":"id","client_secret":"secret","redirect_url":"http://localhost"},"openai":{"api_key":"sk-abc"},"server":{"port":"0","db_url":"postgres://user:pass@localhost/db"}}`
	f, err := os.CreateTemp("", "good_config_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(cfgText)
	f.Close()

	cfg, err := config.LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("unexpected error loading valid config: %v", err)
	}
	r := setupRouter(nil, cfg)
	if r == nil {
		t.Error("expected non-nil router with valid config")
	}
}

func TestServerStartupWithMissingConfig(t *testing.T) {
	_, err := config.LoadConfig("/tmp/definitely-does-not-exist.json")
	if err == nil {
		t.Error("expected error for missing config file, got nil")
	}
}

