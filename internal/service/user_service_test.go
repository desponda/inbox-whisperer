package service

import (
	"context"
	"testing"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/google/uuid"
)

type mockUserRepo struct {
	GetUserFunc        func(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUserFunc     func(ctx context.Context, user *models.User) error
	ListUsersFunc      func(ctx context.Context) ([]*models.User, error)
	UpdateUserFunc     func(ctx context.Context, user *models.User) error
	DeleteFunc         func(ctx context.Context, id uuid.UUID) error
	DeactivateUserFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserRepo) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.GetUserFunc(ctx, id)
}
func (m *mockUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	return m.CreateUserFunc(ctx, user)
}
func (m *mockUserRepo) ListUsers(ctx context.Context) ([]*models.User, error) {
	return m.ListUsersFunc(ctx)
}
func (m *mockUserRepo) UpdateUser(ctx context.Context, user *models.User) error {
	return m.UpdateUserFunc(ctx, user)
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}
func (m *mockUserRepo) DeactivateUser(ctx context.Context, id uuid.UUID) error {
	return m.DeactivateUserFunc(ctx, id)
}

func TestUserService_GetUser(t *testing.T) {
	testUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tests := []struct {
		name     string
		id       uuid.UUID
		repoFunc func(ctx context.Context, id uuid.UUID) (*models.User, error)
		wantUser *models.User
		wantErr  bool
	}{
		{
			name: "success",
			id:   testUUID,
			repoFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{ID: id, Email: "test@example.com"}, nil
			},
			wantUser: &models.User{ID: testUUID, Email: "test@example.com"},
			wantErr:  false,
		},
		{
			name: "repo error",
			id:   testUUID,
			repoFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return nil, context.DeadlineExceeded
			},
			wantUser: nil,
			wantErr:  true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{GetUserFunc: tc.repoFunc}
			var repoIface data.UserRepository = repo
			svc := NewUserService(repoIface)
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
	testUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	t.Run("success", func(t *testing.T) {
		repo := &mockUserRepo{
			ListUsersFunc: func(ctx context.Context) ([]*models.User, error) {
				return []*models.User{{ID: testUUID, Email: "a@example.com"}}, nil
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		users, err := svc.ListUsers(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(users) != 1 || users[0].ID != testUUID {
			t.Errorf("unexpected users: %+v", users)
		}
	})
	t.Run("error", func(t *testing.T) {
		repo := &mockUserRepo{
			ListUsersFunc: func(ctx context.Context) ([]*models.User, error) {
				return nil, context.DeadlineExceeded
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		_, err := svc.ListUsers(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	testUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	t.Run("success", func(t *testing.T) {
		called := false
		repo := &mockUserRepo{
			UpdateUserFunc: func(ctx context.Context, user *models.User) error {
				called = true
				return nil
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.UpdateUser(context.Background(), &models.User{ID: testUUID})
		if err != nil {
			t.Logf("TestUserService_UpdateUser success: input user=%+v, err=%v", &models.User{ID: testUUID}, err)
		}

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected UpdateUserFunc to be called")
		}
	})
	t.Run("error", func(t *testing.T) {
		repo := &mockUserRepo{
			UpdateUserFunc: func(ctx context.Context, user *models.User) error {
				return context.DeadlineExceeded
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.UpdateUser(context.Background(), &models.User{ID: testUUID})
		t.Logf("TestUserService_UpdateUser error: input user=%+v, err=%v", &models.User{ID: testUUID}, err)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	testUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	t.Run("success", func(t *testing.T) {
		called := false
		repo := &mockUserRepo{
			DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
				called = true
				return nil
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.DeleteUser(context.Background(), testUUID)
		t.Logf("TestUserService_DeleteUser success: id=%s, err=%v", testUUID, err)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected DeleteFunc to be called")
		}
	})
	t.Run("error", func(t *testing.T) {
		repo := &mockUserRepo{
			DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
				return context.DeadlineExceeded
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.DeleteUser(context.Background(), testUUID)
		t.Logf("TestUserService_DeleteUser error: id=%s, err=%v", testUUID, err)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUserService_DeactivateUser(t *testing.T) {
	testUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	t.Run("success", func(t *testing.T) {
		called := false
		repo := &mockUserRepo{
			DeactivateUserFunc: func(ctx context.Context, id uuid.UUID) error {
				called = true
				return nil
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.DeactivateUser(context.Background(), testUUID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected DeactivateUserFunc to be called")
		}
	})
	t.Run("error", func(t *testing.T) {
		repo := &mockUserRepo{
			DeactivateUserFunc: func(ctx context.Context, id uuid.UUID) error {
				return context.DeadlineExceeded
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.DeactivateUser(context.Background(), testUUID)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUserService_CreateUser(t *testing.T) {
	testUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	t.Run("success", func(t *testing.T) {
		called := false
		repo := &mockUserRepo{
			CreateUserFunc: func(ctx context.Context, user *models.User) error {
				called = true
				return nil
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.CreateUser(context.Background(), &models.User{ID: testUUID})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected CreateUserFunc to be called")
		}
	})
	t.Run("error", func(t *testing.T) {
		repo := &mockUserRepo{
			CreateUserFunc: func(ctx context.Context, user *models.User) error {
				return context.DeadlineExceeded
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.CreateUser(context.Background(), &models.User{ID: testUUID})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
