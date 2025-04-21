package service

import (
	"context"
	"testing"
	"github.com/desponda/inbox-whisperer/internal/models"
)

type mockUserRepo struct {
	GetByIDFunc func(ctx context.Context, id string) (*models.User, error)
	CreateFunc  func(ctx context.Context, user *models.User) error
	ListFunc    func(ctx context.Context) ([]*models.User, error)
	UpdateFunc  func(ctx context.Context, user *models.User) error
	DeleteFunc  func(ctx context.Context, id string) error
}

func (m *mockUserRepo) List(ctx context.Context) ([]*models.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *models.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	return m.GetByIDFunc(ctx, id)
}
func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	return m.CreateFunc(ctx, user)
}

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		repoFunc  func(ctx context.Context, id string) (*models.User, error)
		wantUser  *models.User
		wantErr   bool
	}{
		{
			name: "success",
			id:   "abc",
			repoFunc: func(ctx context.Context, id string) (*models.User, error) {
				return &models.User{ID: id, Email: "test@example.com"}, nil
			},
			wantUser: &models.User{ID: "abc", Email: "test@example.com"},
			wantErr:  false,
		},
		{
			name: "repo error",
			id:   "xyz",
			repoFunc: func(ctx context.Context, id string) (*models.User, error) {
				return nil, context.DeadlineExceeded
			},
			wantUser: nil,
			wantErr:  true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{GetByIDFunc: tc.repoFunc}
			svc := NewUserService(repo)
			user, err := svc.GetUser(context.Background(), tc.id)
			if (err != nil) != tc.wantErr {
				t.Errorf("expected error=%v, got %v", tc.wantErr, err)
			}
			if tc.wantUser != nil && (user == nil || user.ID != tc.wantUser.ID || user.Email != tc.wantUser.Email) {
				t.Errorf("expected user %+v, got %+v", tc.wantUser, user)
			}
		})
	}
}

func TestUserService_ListUsers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockUserRepo{
			ListFunc: func(ctx context.Context) ([]*models.User, error) {
				return []*models.User{{ID: "a", Email: "a@example.com"}}, nil
			},
		}
		svc := NewUserService(repo)
		users, err := svc.ListUsers(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(users) != 1 || users[0].ID != "a" {
			t.Errorf("unexpected users: %+v", users)
		}
	})
	t.Run("error", func(t *testing.T) {
		repo := &mockUserRepo{
			ListFunc: func(ctx context.Context) ([]*models.User, error) {
				return nil, context.DeadlineExceeded
			},
		}
		svc := NewUserService(repo)
		_, err := svc.ListUsers(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		called := false
		repo := &mockUserRepo{
			UpdateFunc: func(ctx context.Context, user *models.User) error {
				called = true
				return nil
			},
		}
		svc := NewUserService(repo)
		err := svc.UpdateUser(context.Background(), &models.User{ID: "a"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected UpdateFunc to be called")
		}
	})
	t.Run("error", func(t *testing.T) {
		repo := &mockUserRepo{
			UpdateFunc: func(ctx context.Context, user *models.User) error {
				return context.DeadlineExceeded
			},
		}
		svc := NewUserService(repo)
		err := svc.UpdateUser(context.Background(), &models.User{ID: "a"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		called := false
		repo := &mockUserRepo{
			DeleteFunc: func(ctx context.Context, id string) error {
				called = true
				return nil
			},
		}
		svc := NewUserService(repo)
		err := svc.DeleteUser(context.Background(), "a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected DeleteFunc to be called")
		}
	})
	t.Run("error", func(t *testing.T) {
		repo := &mockUserRepo{
			DeleteFunc: func(ctx context.Context, id string) error {
				return context.DeadlineExceeded
			},
		}
		svc := NewUserService(repo)
		err := svc.DeleteUser(context.Background(), "a")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		inputUser *models.User
		repoFunc  func(ctx context.Context, user *models.User) error
		wantErr   bool
		wantCall  bool
	}{
		{
			name:      "success",
			inputUser: &models.User{ID: "abc"},
			repoFunc: func(ctx context.Context, user *models.User) error {
				return nil
			},
			wantErr:  false,
			wantCall: true,
		},
		{
			name:      "repo error",
			inputUser: &models.User{ID: "abc"},
			repoFunc: func(ctx context.Context, user *models.User) error {
				return context.DeadlineExceeded
			},
			wantErr:  true,
			wantCall: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			repo := &mockUserRepo{
				CreateFunc: func(ctx context.Context, user *models.User) error {
					called = true
					return tc.repoFunc(ctx, user)
				},
			}
			svc := NewUserService(repo)
			err := svc.CreateUser(context.Background(), tc.inputUser)
			if (err != nil) != tc.wantErr {
				t.Errorf("expected error=%v, got %v", tc.wantErr, err)
			}
			if called != tc.wantCall {
				t.Errorf("expected called=%v, got %v", tc.wantCall, called)
			}
		})
	}
}
