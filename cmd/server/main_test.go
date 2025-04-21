package main

import (
	"net/http"
	"net/http/httptest"
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
	}
}

