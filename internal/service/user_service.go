package service

import (
	"context"
	"log"
	"errors"
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


func (s *UserService) DeactivateUser(ctx context.Context, id string) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			log.Printf("DeactivateUser: fatal error on GetByID for user %s: %v", id, err)
			return err
		}
		// Only treat 'not found' as idempotent
		if err.Error() == "not found" {
			log.Printf("DeactivateUser: user %s not found (idempotent)", id)
			return nil
		}
		log.Printf("DeactivateUser: unexpected error on GetByID for user %s: %v", id, err)
		return err
	}
	if user.Deactivated {
		log.Printf("DeactivateUser: user %s already deactivated (idempotent)", id)
		return nil // already deactivated
	}
	user.Deactivated = true
	err = s.repo.Update(ctx, user)
	if err != nil {
		log.Printf("DeactivateUser: error updating user %s: %v", id, err)
		return err
	}
	log.Printf("DeactivateUser: user %s deactivated successfully", id)
	return nil
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
