package data

import (
	"context"
	"github.com/desponda/inbox-whisperer/internal/models"
)

// UserRepository defines DB operations for users (interface for service layer)
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	List(ctx context.Context) ([]*models.User, error)
	Update(ctx context.Context, user *models.User) error // supports Deactivated
	Delete(ctx context.Context, id string) error
}

// PostgresUserRepository implements UserRepository for Postgres
// (implements all methods on *DB)
func (db *DB) List(ctx context.Context) ([]*models.User, error) {
	rows, err := db.Pool.Query(ctx, `SELECT id, email, created_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

// Update updates user fields, including Deactivated for soft delete
func (db *DB) Update(ctx context.Context, user *models.User) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE users SET email = $1, deactivated = $2 WHERE id = $3`,
		user.Email, user.Deactivated, user.ID,
	)
	return err
}

func (db *DB) Delete(ctx context.Context, id string) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

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
