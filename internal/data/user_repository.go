package data

import (
	"context"

	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/google/uuid"
)

// UserRepository defines DB operations for users (interface for service layer)
type UserRepository interface {
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	ListUsers(ctx context.Context) ([]*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeactivateUser(ctx context.Context, id uuid.UUID) error
}

// PostgresUserRepository implements UserRepository for Postgres
// (implements all methods on *DB)
func (db *DB) ListUsers(ctx context.Context) ([]*models.User, error) {
	rows, err := db.Pool.Query(ctx, `SELECT id, email, created_at, deactivated FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.Deactivated); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

// Update updates user fields, including Deactivated for soft delete
func (db *DB) UpdateUser(ctx context.Context, user *models.User) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE users SET email = $1, deactivated = $2 WHERE id = $3`,
		user.Email, user.Deactivated, user.ID,
	)
	return err
}

func (db *DB) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (db *DB) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row := db.Pool.QueryRow(ctx, `SELECT id, email, created_at, deactivated FROM users WHERE id = $1`, id)
	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.CreatedAt, &user.Deactivated); err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) CreateUser(ctx context.Context, user *models.User) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO users (id, email, created_at, deactivated)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email`,
		user.ID, user.Email, user.CreatedAt, user.Deactivated,
	)
	return err
}

func (db *DB) DeactivateUser(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `UPDATE users SET deactivated = TRUE WHERE id = $1`, id)
	return err
}
