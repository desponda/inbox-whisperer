package session

import (
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"net/http"

	"github.com/desponda/inbox-whisperer/internal/api/testutils"
)

func setupTestManager(t *testing.T) (*Manager, *testutils.MockStore) {
	// Set up debug logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	store := testutils.NewMockStore()
	manager := &Manager{
		store: store,
	}
	return manager, store
}

func TestManager_Start(t *testing.T) {
	manager, store := setupTestManager(t)

	tests := []struct {
		name          string
		setupCookie   bool
		setupSession  bool
		expectNewSess bool
	}{
		{
			name:          "no cookie, create new session",
			setupCookie:   false,
			setupSession:  false,
			expectNewSess: true,
		},
		{
			name:          "cookie exists but session doesn't",
			setupCookie:   true,
			setupSession:  false,
			expectNewSess: true,
		},
		{
			name:          "cookie and session exist",
			setupCookie:   true,
			setupSession:  true,
			expectNewSess: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			if tc.setupCookie {
				cookieValue := "non-existent-session"
				if tc.setupSession {
					cookieValue = "existing-session"
				}
				r.AddCookie(&http.Cookie{
					Name:  SessionCookieName,
					Value: cookieValue,
				})
			}

			if tc.setupSession {
				sess := testutils.NewMockSession(
					"existing-session",
					"",
					"",
					nil,
					time.Now(),
					time.Now().Add(24*time.Hour),
				)
				err := store.Save(r.Context(), sess)
				require.NoError(t, err)
			}

			// Test Start
			sess, err := manager.Start(w, r)
			require.NoError(t, err)
			assert.NotEmpty(t, sess.ID())
			assert.Empty(t, sess.UserID()) // Initially empty

			// Check cookie
			cookies := w.Result().Cookies()
			if tc.expectNewSess {
				require.NotEmpty(t, cookies, "should have set a cookie")
				require.Equal(t, SessionCookieName, cookies[0].Name)
				require.Equal(t, sess.ID(), cookies[0].Value)
			} else {
				require.Empty(t, cookies, "should not have set a cookie")
			}
		})
	}
}

func TestManager_Destroy(t *testing.T) {
	manager, mockStore := setupTestManager(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Create and save a session first
	sess, err := manager.Start(w, r)
	require.NoError(t, err)

	// Test Destroy
	err = manager.Destroy(w, r)
	require.NoError(t, err)

	// Verify session is gone
	_, err = mockStore.Get(r.Context(), sess.ID())
	assert.Nil(t, err)
}

func TestManager_Refresh(t *testing.T) {
	manager, mockStore := setupTestManager(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Create and save a session first
	sess, err := manager.Start(w, r)
	require.NoError(t, err)

	// Set some data
	sess.SetUserID("test-user-id")
	err = mockStore.Save(r.Context(), sess)
	require.NoError(t, err)

	// Add session cookie to request
	cookies := w.Result().Cookies()
	require.NotEmpty(t, cookies, "session cookie should be set")
	r.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: sess.ID(),
	})

	// Test Refresh
	err = manager.Refresh(w, r)
	require.NoError(t, err)

	// Verify session still exists
	refreshedSess, err := mockStore.Get(r.Context(), sess.ID())
	require.NoError(t, err)
	assert.Equal(t, "test-user-id", refreshedSess.UserID())
}

func TestManager_Store(t *testing.T) {
	manager, mockStore := setupTestManager(t)

	store := manager.Store()
	assert.Equal(t, mockStore, store)
}
