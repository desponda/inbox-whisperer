# Session Management

## Overview

The session management system in Inbox Whisperer uses a hybrid approach:
- Gorilla Sessions for secure cookie management and session handling
- Custom PostgreSQL storage implementation for session persistence
- Clean interface-driven design for flexibility and testability

## Architecture

### Core Interfaces

```go
// Session represents a user session
type Session interface {
    ID() string
    UserID() string
    SetUserID(string)
    CreatedAt() time.Time
    ExpiresAt() time.Time
    Values() map[string]interface{}
    SetValue(key string, value interface{})
    GetValue(key string) (interface{}, bool)
}

// Store manages session persistence
type Store interface {
    Create(ctx context.Context) (Session, error)
    Get(ctx context.Context, id string) (Session, error)
    Save(ctx context.Context, session Session) error
    Delete(ctx context.Context, id string) error
    Cleanup(ctx context.Context) error
}

// Manager handles session lifecycle
type Manager interface {
    Start(w http.ResponseWriter, r *http.Request) (Session, error)
    Destroy(w http.ResponseWriter, r *http.Request) error
    Refresh(w http.ResponseWriter, r *http.Request) error
    Store() Store
}
```

### Implementation

The implementation is located in `internal/auth/service/session/gorilla/` and consists of:

1. `store.go` - PostgreSQL-backed session store
2. `session.go` - Session implementation
3. `manager.go` - Session lifecycle manager

## Usage

### Configuration

```go
config := &gorilla.StoreConfig{
    DB:            db,                    // *sql.DB instance
    TableName:     "sessions",            // PostgreSQL table name
    SessionName:   "my_app_session",      // Cookie name
    AuthKey:       []byte("auth-key"),    // Cookie authentication key
    EncryptionKey: []byte("enc-key"),     // Optional encryption key
    Path:          "/",                   // Cookie path
    Domain:        "",                    // Cookie domain (empty for current domain)
    MaxAge:        86400,                 // Session lifetime in seconds
    Secure:        true,                  // Require HTTPS
    HttpOnly:      true,                  // Prevent JavaScript access
    SameSite:      http.SameSiteStrict,   // CSRF protection
}

store, err := gorilla.NewStore(config)
if err != nil {
    log.Fatal(err)
}

manager := gorilla.NewManager(store, config.SessionName)
```

### Middleware

```go
func SessionMiddleware(manager interfaces.Manager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            session, err := manager.Start(w, r)
            if err != nil {
                http.Error(w, "Session error", http.StatusInternalServerError)
                return
            }

            // Store session in context
            ctx := interfaces.WithSession(r.Context(), session)
            next.ServeHTTP(w, r.WithContext(ctx))

            // Refresh session
            if err := manager.Refresh(w, r); err != nil {
                // Log error but don't fail request
                log.Printf("Failed to refresh session: %v", err)
            }
        })
    }
}
```

### Handler Example

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    session, ok := interfaces.GetSession(r.Context())
    if !ok {
        http.Error(w, "No session", http.StatusUnauthorized)
        return
    }

    // Get value
    if value, ok := session.GetValue("my_key"); ok {
        // Use value
    }

    // Set value
    session.SetValue("my_key", "my_value")
}
```

## Security Features

1. **Secure Cookies**
   - Signed using HMAC
   - Optional encryption
   - HTTP-only
   - Secure-only (HTTPS)
   - SameSite protection

2. **Session Security**
   - Server-side storage
   - Automatic cleanup of expired sessions
   - Session ID using UUID v4
   - Session fixation protection

3. **CSRF Protection**
   - SameSite cookie attribute
   - Secure session ID generation
   - No session ID in URL

## Database Schema

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    values JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);
```

## Testing

The interface-driven design makes testing straightforward:

1. **Unit Tests**: Use mock implementations of the interfaces
2. **Integration Tests**: Use test database with proper schema
3. **End-to-End Tests**: Test full session lifecycle

## Future Improvements

1. Configurable session duration
2. Redis support for distributed environments
3. Session activity tracking
4. Rate limiting for session operations
5. Enhanced logging and monitoring 