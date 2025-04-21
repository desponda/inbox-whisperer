package data

import (
	"context"
	"github.com/desponda/inbox-whisperer/internal/models"
)

// UserRepository defines DB operations for users (interface for service layer)
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}

// PostgresUserRepository implements UserRepository for Postgres
// (implements all methods on *DB)
func (db *DB) GetByID(ctx context.Context, id string) (*models.User, error) {
	row := db.Pool.QueryRow(ctx, `SELECT id, email, created_at FROM users WHERE id = $1`, id)
	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) Create(ctx context.Context, user *models.User) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO users (id, email, created_at) VALUES ($1, $2, $3)`,
		user.ID, user.Email, user.CreatedAt,
	)
	return err
}
