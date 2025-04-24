package data

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestDB_New_and_Close(t *testing.T) {
	// Use a known-bad URL to test error path
	badURL := "postgres://invalid:invalid@localhost:5432/invalid"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db, err := New(ctx, badURL)
	if err == nil {
		t.Error("expected error for bad DB URL, got nil")
	}
	if db != nil {
		t.Error("expected nil db for bad DB URL")
	}

	// Use a valid URL if available in env, otherwise skip success test
	goodURL := testDBURLFromEnv()
	if goodURL == "" {
		t.Skip("No valid DB URL in env; skipping DB connection success test")
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	db, err = New(ctx2, goodURL)
	if err != nil {
		t.Errorf("expected no error for valid DB URL, got %v", err)
	}
	if db == nil || db.Pool == nil {
		t.Error("expected non-nil db and db.Pool for valid DB URL")
	}
	if db != nil {
		db.Close()
	}
}

func testDBURLFromEnv() string {
	// Try standard env vars, return empty string if not set
	for _, env := range []string{"DATABASE_URL", "TEST_DATABASE_URL"} {
		if url := getenv(env); url != "" {
			return url
		}
	}
	return ""
}

func getenv(key string) string {
	if v, ok := syscallEnv(key); ok {
		return v
	}
	return ""
}

// syscallEnv is separated for testability
var syscallEnv = func(key string) (string, bool) {
	return os.LookupEnv(key)
}
