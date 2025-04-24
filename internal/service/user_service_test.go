package service

import (
	"context"
	"errors"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"testing"
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
		name     string
		id       string
		repoFunc func(ctx context.Context, id string) (*models.User, error)
		wantUser *models.User
		wantErr  bool
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
	t.Run("success", func(t *testing.T) {
		repo := &mockUserRepo{
			ListFunc: func(ctx context.Context) ([]*models.User, error) {
				return []*models.User{{ID: "a", Email: "a@example.com"}}, nil
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
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
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
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
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.UpdateUser(context.Background(), &models.User{ID: "a"})
		if err != nil {
			t.Logf("TestUserService_UpdateUser success: input user=%+v, err=%v", &models.User{ID: "a"}, err)
		}

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
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.UpdateUser(context.Background(), &models.User{ID: "a"})
		t.Logf("TestUserService_UpdateUser error: input user=%+v, err=%v", &models.User{ID: "a"}, err)
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
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.DeleteUser(context.Background(), "a")
		t.Logf("TestUserService_DeleteUser success: id=%s, err=%v", "a", err)
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
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.DeleteUser(context.Background(), "a")
		t.Logf("TestUserService_DeleteUser error: id=%s, err=%v", "a", err)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUserService_DeactivateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		user := &models.User{ID: "abc", Deactivated: false}
		getByIDCalled := false
		updateCalled := false
		repo := &mockUserRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
				getByIDCalled = true
				return user, nil
			},
			UpdateFunc: func(ctx context.Context, u *models.User) error {
				updateCalled = true
				if !u.Deactivated {
					t.Errorf("expected user to be deactivated, got %+v", u)
				}
				return nil
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.DeactivateUser(context.Background(), "abc")
		if err != nil {
			t.Logf("TestUserService_DeactivateUser success: id=%s, err=%v", "abc", err)
		}

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !getByIDCalled || !updateCalled {
			t.Error("expected GetByID and Update to be called")
		}
	})

	t.Run("already deactivated (idempotent)", func(t *testing.T) {
		user := &models.User{ID: "abc", Deactivated: true}
		getByIDCalled := false
		updateCalled := false
		repo := &mockUserRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
				getByIDCalled = true
				return user, nil
			},
			UpdateFunc: func(ctx context.Context, u *models.User) error {
				updateCalled = true
				return nil
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.DeactivateUser(context.Background(), "abc")
		if err != nil {
			t.Logf("TestUserService_DeactivateUser already deactivated: id=%s, err=%v", "abc", err)
		}

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !getByIDCalled {
			t.Error("expected GetByID to be called")
		}
		if updateCalled {
			t.Error("expected Update NOT to be called for already deactivated user")
		}
	})

	t.Run("user not found (idempotent)", func(t *testing.T) {
		getByIDCalled := false
		repo := &mockUserRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
				getByIDCalled = true
				return nil, errors.New("not found")
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.DeactivateUser(context.Background(), "no-user")
		if err != nil {
			t.Logf("TestUserService_DeactivateUser not found: id=%s, err=%v", "no-user", err)
		}

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !getByIDCalled {
			t.Error("expected GetByID to be called")
		}
	})

	t.Run("repo error on GetByID", func(t *testing.T) {
		repo := &mockUserRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
				return nil, context.DeadlineExceeded
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.DeactivateUser(context.Background(), "abc")
		if err != nil {
			t.Logf("TestUserService_DeactivateUser repo error GetByID: id=%s, err=%v", "abc", err)
		}

		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("repo error on Update", func(t *testing.T) {
		user := &models.User{ID: "abc", Deactivated: false}
		repo := &mockUserRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
				return user, nil
			},
			UpdateFunc: func(ctx context.Context, u *models.User) error {
				return context.DeadlineExceeded
			},
		}
		var repoIface data.UserRepository = repo
		svc := NewUserService(repoIface)
		err := svc.DeactivateUser(context.Background(), "abc")
		if err != nil {
			t.Logf("TestUserService_DeactivateUser repo error Update: id=%s, err=%v", "abc", err)
		}

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
			var repoIface data.UserRepository = repo
			svc := NewUserService(repoIface)
			err := svc.CreateUser(context.Background(), tc.inputUser)
			if err != nil {
				t.Logf("TestUserService_CreateUser error: input user=%+v, err=%v", tc.inputUser, err)
			}

			if (err != nil) != tc.wantErr {
				t.Errorf("expected error=%v, got %v", tc.wantErr, err)
			}
			if called != tc.wantCall {
				t.Errorf("expected called=%v, got %v", tc.wantCall, called)
			}
		})
	}
}
