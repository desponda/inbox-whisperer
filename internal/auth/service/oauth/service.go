package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
	goauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// Common errors
var (
	ErrTokenExchange  = errors.New("failed to exchange authorization code for token")
	ErrUserInfo       = errors.New("failed to retrieve user information")
	ErrInvalidEmail   = errors.New("invalid email format")
	ErrInvalidIDToken = errors.New("invalid ID token")
)

type Service struct {
	config           *oauth2.Config
	userTokens       data.UserTokenRepository
	userRepo         data.UserRepository
	userIdentityRepo data.UserIdentityRepository
}

func NewService(config *oauth2.Config, userTokens data.UserTokenRepository, userRepo data.UserRepository, userIdentityRepo data.UserIdentityRepository) *Service {
	return &Service{
		config:           config,
		userTokens:       userTokens,
		userRepo:         userRepo,
		userIdentityRepo: userIdentityRepo,
	}
}

// GetDB returns the underlying database interface
func (s *Service) GetDB() interface{} {
	if db, ok := s.userTokens.(interface{ GetDB() interface{} }); ok {
		return db.GetDB()
	}
	return nil
}

// ExchangeCodeForToken exchanges the OAuth code for tokens and validates them
func (s *Service) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, *goauth2.Userinfo, error) {
	tok, err := s.config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrTokenExchange, err)
	}

	if tok.AccessToken == "" {
		return nil, nil, fmt.Errorf("%w: empty access token", ErrTokenExchange)
	}

	// Extract and validate ID token
	idToken, ok := tok.Extra("id_token").(string)
	if !ok || idToken == "" {
		return nil, nil, fmt.Errorf("%w: missing ID token", ErrTokenExchange)
	}

	// Validate ID token
	if err := s.validateIDToken(ctx, idToken); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidIDToken, err)
	}

	// Get user info using Google's official client
	oauth2Service, err := goauth2.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(tok)))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to create OAuth2 service: %v", ErrUserInfo, err)
	}

	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrUserInfo, err)
	}

	// Validate email
	if userInfo.Email == "" {
		return nil, nil, fmt.Errorf("%w: missing email", ErrUserInfo)
	}
	if _, err := mail.ParseAddress(userInfo.Email); err != nil {
		return nil, nil, ErrInvalidEmail
	}

	return tok, userInfo, nil
}

// SaveUserAndToken saves the user and token information using repositories
func (s *Service) SaveUserAndToken(ctx context.Context, userInfo *goauth2.Userinfo, tok *oauth2.Token) error {
	const provider = "google"

	// 1. Look up user_identity by provider and provider_user_id
	identity, err := s.userIdentityRepo.GetByProviderAndProviderUserID(ctx, provider, userInfo.Id)
	var userID uuid.UUID
	if err != nil {
		// Not found: create new user and user_identity
		userID = uuid.New()
		user := &models.User{
			ID:        userID,
			Email:     userInfo.Email,
			CreatedAt: time.Now().UTC(),
		}
		if err := s.userRepo.CreateUser(ctx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		identity = &models.UserIdentity{
			ID:             uuid.New(),
			UserID:         userID,
			Provider:       provider,
			ProviderUserID: userInfo.Id,
			Email:          userInfo.Email,
			CreatedAt:      time.Now().UTC(),
		}
		if err := s.userIdentityRepo.Create(ctx, identity); err != nil {
			return fmt.Errorf("failed to create user identity: %w", err)
		}
	} else {
		userID = identity.UserID
	}

	// 2. Save user token
	if err := s.userTokens.SaveUserToken(ctx, userID.String(), tok); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// Internal methods moved from auth.go...
func (s *Service) validateIDToken(ctx context.Context, tokenString string) error {
	_, err := idtoken.Validate(ctx, tokenString, s.config.ClientID)
	if err != nil {
		return fmt.Errorf("failed to validate ID token: %w", err)
	}
	return nil
}
