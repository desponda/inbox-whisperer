package service

import (
	"context"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
)

// UserService provides business logic for users
type UserServiceInterface interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	ListUsers(ctx context.Context) ([]*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	DeactivateUser(ctx context.Context, id string) error
}

type UserService struct {
	repo data.UserRepository
}

func (s *UserService) ListUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	return s.repo.Update(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// DeactivateUser performs a soft delete (sets Deactivated=true)
func (s *UserService) DeactivateUser(ctx context.Context, id string) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		// If user not found, treat as successful delete (idempotent)
		return nil
	}
	if user.Deactivated {
		return nil // already deactivated
	}
	user.Deactivated = true
	return s.repo.Update(ctx, user)
}

func NewUserService(repo data.UserRepository) UserServiceInterface {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	return s.repo.Create(ctx, user)
}
