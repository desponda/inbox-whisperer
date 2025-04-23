package data

import (
	"context"
	"encoding/json"
	"time"

	"golang.org/x/oauth2"
)

type UserTokenRepository interface {
	SaveUserToken(ctx context.Context, userID string, token *oauth2.Token) error
	GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error)
}



func (db *DB) SaveUserToken(ctx context.Context, userID string, token *oauth2.Token) error {
	tokBytes, err := json.Marshal(token)
	if err != nil {
		return err
	}
	_, err = db.Pool.Exec(ctx, `INSERT INTO user_tokens (user_id, token_json, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET token_json = $2, updated_at = $3`,
		userID, string(tokBytes), time.Now().UTC(),
	)
	return err
}

func (db *DB) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	row := db.Pool.QueryRow(ctx, `SELECT token_json FROM user_tokens WHERE user_id = $1`, userID)
	var tokenJSON string
	if err := row.Scan(&tokenJSON); err != nil {
		return nil, err
	}
	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil, err
	}
	return &token, nil
}
