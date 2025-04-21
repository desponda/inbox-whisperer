package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

)

func TestHealthz(t *testing.T) {
	r := setupRouter(nil)
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

