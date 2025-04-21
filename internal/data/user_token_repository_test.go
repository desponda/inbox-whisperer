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
	userID := "user_123"

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
