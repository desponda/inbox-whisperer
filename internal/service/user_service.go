package service

import (
	"context"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/google/uuid"
)

// UserService provides business logic for users
type UserServiceInterface interface {
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	ListUsers(ctx context.Context) ([]*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	DeactivateUser(ctx context.Context, id uuid.UUID) error
}

type UserService struct {
	repo data.UserRepository
}

func (s *UserService) ListUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.ListUsers(ctx)
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	return s.repo.UpdateUser(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) DeactivateUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeactivateUser(ctx, id)
}

func NewUserService(repo data.UserRepository) UserServiceInterface {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	return s.repo.CreateUser(ctx, user)
}
