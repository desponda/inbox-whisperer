package service

import (
	"context"
	"testing"
	"github.com/desponda/inbox-whisperer/internal/models"
)

type mockUserRepo struct {
	GetByIDFunc func(ctx context.Context, id string) (*models.User, error)
	CreateFunc  func(ctx context.Context, user *models.User) error
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	return m.CreateFunc(ctx, user)
}

func TestUserService_GetUser(t *testing.T) {
	repo := &mockUserRepo{
		GetByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
			return &models.User{ID: id, Email: "test@example.com"}, nil
		},
	}
	svc := NewUserService(repo)
	user, err := svc.GetUser(context.Background(), "abc")
	if err != nil || user.ID != "abc" {
		t.Fatalf("expected user, got %v, err %v", user, err)
	}
}

func TestUserService_CreateUser(t *testing.T) {
	called := false
	repo := &mockUserRepo{
		CreateFunc: func(ctx context.Context, user *models.User) error {
			called = true
			return nil
		},
	}
	svc := NewUserService(repo)
	err := svc.CreateUser(context.Background(), &models.User{ID: "abc"})
	if err != nil || !called {
		t.Fatalf("expected Create to be called, got err %v", err)
	}
}
