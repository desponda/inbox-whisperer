package data

import (
	"context"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestUserTokenRepository_SaveAndGet(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	tok := &oauth2.Token{
		AccessToken:  "abc",
		RefreshToken: "def",
		Expiry:       time.Now().Add(1 * time.Hour).UTC(),
		TokenType:    "Bearer",
	}
	userID := "11111111-1111-1111-1111-111111111111"

	db.Pool.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, userID, "test@example.com")

	if err := db.SaveUserToken(ctx, userID, tok); err != nil {
		t.Fatalf("SaveUserToken failed: %v", err)
	}

	got, err := db.GetUserToken(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserToken failed: %v", err)
	}
	if got.AccessToken != tok.AccessToken || got.RefreshToken != tok.RefreshToken || got.TokenType != tok.TokenType {
		t.Errorf("got %+v, want %+v", got, tok)
	}
}
