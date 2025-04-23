package service

import (
	"context"
	"errors"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/rs/zerolog/log"
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
			log.Error().Str("userID", id).Err(err).Msg("DeactivateUser: fatal error on GetByID")
			return err
		}
		// Only treat 'not found' as idempotent
		if err.Error() == "not found" {
			log.Warn().Str("userID", id).Msg("DeactivateUser: user not found (idempotent)")
			return nil
		}
		log.Error().Str("userID", id).Err(err).Msg("DeactivateUser: unexpected error on GetByID")
		return err
	}
	if user.Deactivated {
		log.Warn().Str("userID", id).Msg("DeactivateUser: already deactivated (idempotent)")
		return nil // already deactivated
	}
	user.Deactivated = true
	err = s.repo.Update(ctx, user)
	if err != nil {
		log.Error().Str("userID", id).Err(err).Msg("DeactivateUser: error updating user")
		return err
	}
	log.Info().Str("userID", id).Msg("DeactivateUser: user deactivated successfully")
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
