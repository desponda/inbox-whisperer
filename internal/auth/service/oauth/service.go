package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	goauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

const (
	defaultTimeout = 10 * time.Second
	googleJWKSURL  = "https://www.googleapis.com/oauth2/v3/certs"
)

// Common errors
var (
	ErrTokenExchange  = errors.New("failed to exchange authorization code for token")
	ErrUserInfo       = errors.New("failed to retrieve user information")
	ErrInvalidEmail   = errors.New("invalid email format")
	ErrInvalidIDToken = errors.New("invalid ID token")
)

type Service struct {
	config     *oauth2.Config
	userTokens data.UserTokenRepository
	log        zerolog.Logger
}

func NewService(config *oauth2.Config, userTokens data.UserTokenRepository) *Service {
	return &Service{
		config:     config,
		userTokens: userTokens,
		log:        zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger(),
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

// SaveUserAndToken saves the user and token information in a transaction
func (s *Service) SaveUserAndToken(ctx context.Context, userInfo *goauth2.Userinfo, tok *oauth2.Token) error {
	// Get the underlying pgx.Pool from the repository
	db, ok := s.userTokens.(interface {
		GetDB() *pgxpool.Pool
	})
	if !ok {
		return errors.New("user token repository does not support transactions")
	}

	// Start a transaction
	tx, err := db.GetDB().Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Look up user_identity by provider and provider_user_id
	var userID uuid.UUID
	const provider = "google"
	row := tx.QueryRow(ctx, `SELECT user_id FROM user_identities WHERE provider = $1 AND provider_user_id = $2`, provider, userInfo.Id)
	err = row.Scan(&userID)
	if err == pgx.ErrNoRows {
		// Not found: create new user and user_identity
		userID = uuid.New()
		user := &models.User{
			ID:        userID,
			Email:     userInfo.Email,
			CreatedAt: time.Now().UTC(),
		}
		if err := s.createUserIfNotExists(ctx, tx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		identity := &models.UserIdentity{
			ID:             uuid.New(),
			UserID:         userID,
			Provider:       provider,
			ProviderUserID: userInfo.Id,
			Email:          userInfo.Email,
			CreatedAt:      time.Now().UTC(),
		}
		if err := s.createUserIdentity(ctx, tx, identity); err != nil {
			return fmt.Errorf("failed to create user identity: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to look up user identity: %w", err)
	}
	// else: userID is set from lookup

	// 2. Save user token
	if err := s.saveUserToken(ctx, tx, userID, tok); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Internal methods moved from auth.go...
func (s *Service) validateIDToken(ctx context.Context, tokenString string) error {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return errors.New("missing key ID in token header")
	}

	key, err := s.getGooglePublicKey(ctx, kid)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return ErrInvalidIDToken
	}

	return s.validateTokenClaims(claims)
}

func (s *Service) validateTokenClaims(claims jwt.MapClaims) error {
	iss, ok := claims["iss"].(string)
	if !ok || !strings.HasPrefix(iss, "https://accounts.google.com") {
		return errors.New("invalid token issuer")
	}

	aud, ok := claims["aud"].(string)
	if !ok || aud != s.config.ClientID {
		return errors.New("invalid token audience")
	}

	exp, ok := claims["exp"].(float64)
	if !ok || float64(time.Now().Unix()) > exp {
		return errors.New("token has expired")
	}

	iat, ok := claims["iat"].(float64)
	if !ok || float64(time.Now().Unix()) < iat {
		return errors.New("token used before issued")
	}

	return nil
}

// Private helper methods...
func (s *Service) createUserIfNotExists(ctx context.Context, tx pgx.Tx, user *models.User) error {
	query := `
		INSERT INTO users (id, email, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE
		SET email = EXCLUDED.email
	`
	_, err := tx.Exec(ctx, query, user.ID, user.Email, user.CreatedAt)
	return err
}

func (s *Service) saveUserToken(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tok *oauth2.Token) error {
	query := `
		INSERT INTO user_tokens (user_id, token_json, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET token_json = EXCLUDED.token_json,
			updated_at = EXCLUDED.updated_at
	`
	tokenJSON, err := json.Marshal(tok)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, query, userID, string(tokenJSON), time.Now().UTC())
	return err
}

func (s *Service) createUserIdentity(ctx context.Context, tx pgx.Tx, identity *models.UserIdentity) error {
	query := `
		INSERT INTO user_identities (id, user_id, provider, provider_user_id, email, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (provider, provider_user_id) DO UPDATE
		SET email = EXCLUDED.email
	`
	_, err := tx.Exec(ctx, query, identity.ID, identity.UserID, identity.Provider, identity.ProviderUserID, identity.Email, identity.CreatedAt)
	return err
}

func (s *Service) getGooglePublicKey(ctx context.Context, kid string) (interface{}, error) {
	set, err := jwk.Fetch(ctx, googleJWKSURL)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to fetch Google JWKS with jwx/jwk")
		return nil, err
	}

	key, ok := set.LookupKeyID(kid)
	if !ok {
		s.log.Error().Str("kid", kid).Msg("Requested key ID not found in Google's JWKS (jwx/jwk)")
		return nil, errors.New("key not found")
	}

	rawKey, err := jwk.PublicKeyOf(key)
	if err != nil {
		s.log.Error().Err(err).Str("kid", kid).Msg("Failed to get raw public key from JWK")
		return nil, errors.New("failed to get raw public key from JWK")
	}
	return rawKey, nil
}
