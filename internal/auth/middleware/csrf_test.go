package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/api/testutils"
	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestCSRF() (*CSRF, session.Manager, func()) {
	// Create a mock store
	store := testutils.NewMockStore()

	// Create a mock session manager that reuses sessions
	var currentSession session.Session
	manager := testutils.NewMockSessionManager(func(w http.ResponseWriter, r *http.Request) (session.Session, error) {
		// Check for existing session cookie
		if cookie, err := r.Cookie("session_id"); err == nil {
			if sess, err := store.Get(r.Context(), cookie.Value); err == nil {
				return sess, nil
			}
		}

		// Create a new session if none exists
		if currentSession == nil {
			sess := testutils.NewMockSession(
				"test-session",
				"user1",
				"", // role
				make(map[string]interface{}),
				time.Now(),
				time.Now().Add(24*time.Hour),
			)
			currentSession = sess

			// Set the session cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sess.ID(),
				Path:     "/",
				HttpOnly: true,
				Secure:   false, // For testing
				Expires:  sess.ExpiresAt(),
			})

			// Save the session
			err := store.Save(r.Context(), sess)
			if err != nil {
				return nil, err
			}
		}

		return currentSession, nil
	})

	// Set the store function
	manager.StoreFunc = func() session.Store {
		return store
	}

	csrf := NewCSRF(manager)
	cleanup := func() {
		currentSession = nil
	}

	return csrf, manager, cleanup
}

func TestCSRF_SafeMethods(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	csrf, _, cleanup := setupTestCSRF()
	defer cleanup()

	handler := csrf.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	// Test GET request
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)

	// Should succeed and generate a token
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())

	// Get the session cookie
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)

	// Make another request with the cookie
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.AddCookie(cookies[0])
	handler.ServeHTTP(w2, r2)

	// Should succeed and reuse the same token
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "ok", w2.Body.String())
}

func TestCSRF_UnsafeMethods(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	csrf, manager, cleanup := setupTestCSRF()
	defer cleanup()

	handler := csrf.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	// First, make a GET request to get a token
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)

	// Get the session cookie and token
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1, "Should have a session cookie")
	sessionCookie := cookies[0]
	t.Logf("Session cookie: %+v", sessionCookie)

	// Get the session to extract the token
	session, err := manager.Start(w, r)
	require.NoError(t, err, "Should be able to start session")
	t.Logf("Session ID: %s", session.ID())
	t.Logf("Session values: %+v", session.Values())

	token, ok := session.GetValue(CSRFTokenKey)
	require.True(t, ok, "CSRF token should be set in session")
	require.NotEmpty(t, token, "CSRF token should not be empty")
	t.Logf("CSRF token: %v", token)

	// Verify the token is actually saved in the store
	storedSession, err := manager.Store().Get(r.Context(), session.ID())
	require.NoError(t, err, "Should be able to get session from store")
	t.Logf("Stored session values: %+v", storedSession.Values())
	storedToken, ok := storedSession.GetValue(CSRFTokenKey)
	require.True(t, ok, "CSRF token should be in stored session")
	require.Equal(t, token, storedToken, "Stored token should match session token")

	tests := []struct {
		name       string
		method     string
		token      string
		useForm    bool
		wantStatus int
	}{
		{
			name:       "valid token in header",
			method:     "POST",
			token:      token.(string),
			useForm:    false,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing token",
			method:     "POST",
			token:      "",
			useForm:    false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid token",
			method:     "POST",
			token:      "invalid",
			useForm:    false,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "valid token in form",
			method:     "POST",
			token:      token.(string),
			useForm:    true,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var r *http.Request

			if tt.useForm {
				form := url.Values{}
				form.Set(CSRFFormField, tt.token)
				body := strings.NewReader(form.Encode())
				r = httptest.NewRequest(tt.method, "/", body)
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				r = httptest.NewRequest(tt.method, "/", nil)
				if tt.token != "" {
					r.Header.Set(CSRFHeaderName, tt.token)
				}
			}
			r.AddCookie(sessionCookie)

			handler.ServeHTTP(w, r)
			assert.Equal(t, tt.wantStatus, w.Code, "unexpected status code for %s", tt.name)
			t.Logf("%s response: %d", tt.name, w.Code)
		})
	}
}

func TestCSRF_TokenGeneration(t *testing.T) {
	// Test that tokens are unique
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateCSRFToken()
		require.NoError(t, err)
		assert.False(t, tokens[token], "duplicate token generated")
		tokens[token] = true
	}
}

func TestCSRF_TokenComparison(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "identical tokens",
			a:    "token123",
			b:    "token123",
			want: true,
		},
		{
			name: "different tokens same length",
			a:    "token123",
			b:    "token456",
			want: false,
		},
		{
			name: "different lengths",
			a:    "token123",
			b:    "token1234",
			want: false,
		},
		{
			name: "empty tokens",
			a:    "",
			b:    "",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareTokens(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}
