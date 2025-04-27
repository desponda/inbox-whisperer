# Session Management Refactor

## Overview
Implement a clean, maintainable, and secure session management system using dependency injection and interface-driven design. The initial implementation will use PostgreSQL for storage, but the architecture will support easy switching to other backends like Redis in the future.

## Project Structure
```
internal/
├── auth/                      # Authentication domain
│   ├── interfaces/            # Core auth interfaces
│   │   ├── session.go         # Session interfaces
│   │   └── oauth.go           # OAuth interfaces
│   ├── models/                # Auth domain models
│   │   ├── session.go         # Session entity
│   │   ├── claims.go          # JWT claims
│   │   └── identity.go        # User identity model
│   ├── service/              # Auth business logic
│   │   ├── session/          # Session management
│   │   │   └── gorilla/      # Gorilla-based implementation
│   │   └── oauth/            # OAuth providers
│   │       └── google/       # Google implementation
│   ├── middleware/           # Auth middleware
│   │   ├── auth.go           # Authentication middleware
│   │   └── session.go        # Session middleware
│   └── config/              # Auth configuration
├── api/                     # HTTP handlers
│   └── auth/                # Auth-related handlers
├── data/                   # Data access layer
│   ├── session_repository.go  # Session storage implementation
│   ├── auth_repository.go     # Auth-related data access
│   └── db.go                 # Database connection
└── service/                # Business logic services
```

### Directory Responsibilities

#### `auth/`
Centralizes all authentication and authorization concerns:
- Session management
- OAuth authentication
- JWT handling
- Auth middleware
- User identity

#### `auth/interfaces/`
Defines core auth domain interfaces:
```go
// session.go
type SessionStore interface { ... }
type SessionManager interface { ... }

// oauth.go
type OAuthProvider interface { ... }
type IdentityProvider interface { ... }
```

#### `auth/models/`
Contains auth domain models:
```go
// identity.go
type UserIdentity struct {
    UserID         string
    Provider       string
    ProviderUserID string
    Email          string
    CreatedAt      time.Time
}

// session.go
type Session struct { ... }

// claims.go
type JWTClaims struct { ... }
```

#### `auth/session/`
Implements session management:
- Store implementations (PostgreSQL, Redis)
- Session manager (Gorilla-based)
- Session utilities

#### `auth/oauth/`
Handles OAuth authentication:
- Provider interfaces
- Google implementation
- Future providers (GitHub, etc.)

## Core Interfaces

### Session Store Interface
```go
type SessionStore interface {
    Create(ctx context.Context, session *models.Session) error
    Get(ctx context.Context, id string) (*models.Session, error)
    Update(ctx context.Context, session *models.Session) error
    Delete(ctx context.Context, id string) error
    Cleanup(ctx context.Context) error
}
```

### Session Manager Interface
```go
type SessionManager interface {
    StartSession(ctx context.Context, w http.ResponseWriter, userID string) (*models.Session, error)
    GetSession(r *http.Request) (*models.Session, error)
    RefreshSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error
    EndSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}
```

## Domain Models

### Session Model
```go
type Session struct {
    ID        string
    UserID    string
    Data      map[string]interface{}
    CreatedAt time.Time
    ExpiresAt time.Time
    LastAccessed time.Time
}
```

## Database Schema
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users(id),
    data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Dependency Injection Setup
```go
// Factory for creating session stores
type SessionStoreFactory interface {
    CreateStore(config *Config) (SessionStore, error)
}

// Application container
type Container struct {
    SessionStore    SessionStore
    SessionManager  SessionManager
    Config         *Config
}

// Wire it together
func NewContainer(config *Config) *Container {
    store := postgres.NewSessionStore(config.DB)
    manager := gorilla.NewSessionManager(store, config)
    return &Container{
        SessionStore: store,
        SessionManager: manager,
        Config: config,
    }
}
```

## Implementation Plan

### Phase 1: Core Infrastructure
1. Define interfaces and models
2. Create PostgreSQL schema
3. Implement basic error types
4. Set up dependency injection container

### Phase 2: PostgreSQL Implementation
1. Implement PostgreSQL session store
2. Add database migrations
3. Add store tests
4. Implement cleanup job

### Phase 3: Session Management
1. Implement Gorilla-based session manager
2. Add secure cookie handling
3. Implement session middleware
4. Add manager tests

### Phase 4: Security Features
1. Add CSRF protection
2. Implement rate limiting
3. Add security headers
4. Add session rotation
5. Implement audit logging

### Phase 5: Frontend Integration
1. Update auth context
2. Add token refresh
3. Improve error handling
4. Add integration tests

## Security Considerations

### Session ID Generation
```go
type IDGenerator interface {
    GenerateID() string
}

type UUIDGenerator struct{}

func (g *UUIDGenerator) GenerateID() string {
    return uuid.New().String()
}
```

### Cookie Settings
```go
type CookieConfig struct {
    Name     string
    Path     string
    Domain   string
    MaxAge   int
    Secure   bool
    HttpOnly bool
    SameSite http.SameSite
}
```

## Testing Strategy

### Unit Tests
1. Interface mocks for each component
2. Table-driven tests for core logic
3. Boundary testing for security features

### Integration Tests
1. Database integration tests
2. HTTP middleware tests
3. End-to-end session flow tests

### Security Tests
1. Session fixation tests
2. CSRF protection tests
3. Cookie security tests

## Monitoring and Observability

### Metrics
1. Session creation/deletion rate
2. Session duration
3. Error rates by type
4. Database operation latencies

### Logging
```go
type SessionLogger interface {
    LogSessionCreated(ctx context.Context, sessionID string)
    LogSessionAccessed(ctx context.Context, sessionID string)
    LogSessionError(ctx context.Context, sessionID string, err error)
}
```

## Future Extensibility

### Redis Implementation
```go
type RedisSessionStore struct {
    client *redis.Client
    config *Config
}

func NewRedisSessionStore(client *redis.Client, config *Config) SessionStore {
    return &RedisSessionStore{
        client: client,
        config: config,
    }
}
```

### Distributed Session Support
```go
type DistributedSessionManager interface {
    SessionManager
    Lock(ctx context.Context, sessionID string) error
    Unlock(ctx context.Context, sessionID string) error
}
```

## Code Cleanup Strategy

### Current Code Analysis
1. Review and document existing session code:
   - `/internal/session/session.go`
   - Related middleware in `/internal/api/`
   - Any session-related tests

2. Identify test coverage gaps:
   - Unit tests for session management
   - Integration tests for auth flows
   - End-to-end session scenarios

### Cleanup Tasks
1. Create new tests before refactoring:
   - Session middleware tests
   - Auth flow integration tests
   - Database interaction tests

2. Remove deprecated code:
   - Move existing session.go logic to new structure
   - Update import paths in dependent files
   - Remove old session package
   - Update middleware to use new implementation

3. Update documentation:
   - Remove old session references
   - Update API documentation
   - Update OpenAPI spec for session endpoints

## Implementation Plan

### Phase 1: Setup & Testing
1. Create new auth package structure
2. Implement session repository with tests
3. Implement session service with tests
4. Set up database schema
5. Create integration test suite

## Testing Strategy

### Unit Tests
1. Session repository
   ```go
   func TestSessionRepository_Create(t *testing.T)
   func TestSessionRepository_Get(t *testing.T)
   func TestSessionRepository_Update(t *testing.T)
   func TestSessionRepository_Delete(t *testing.T)
   ```

2. Session service
   ```go
   func TestSessionService_StartSession(t *testing.T)
   func TestSessionService_ValidateSession(t *testing.T)
   func TestSessionService_RefreshSession(t *testing.T)
   ```

### Integration Tests
1. Auth flow tests
   ```go
   func TestAuthFlow_LoginToLogout(t *testing.T)
   func TestAuthFlow_SessionExpiry(t *testing.T)
   func TestAuthFlow_ConcurrentSessions(t *testing.T)
   ```

2. Database tests
   ```go
   func TestSessionRepository_Integration(t *testing.T)
   func TestSessionCleanup_Integration(t *testing.T)
   ```

## Progress Tracking

### Phase 1: Initial Setup & Analysis
- [ ] Review existing session implementation
  - [ ] Document current session.go functionality
  - [ ] Map out dependencies and imports
  - [ ] Identify test coverage gaps
- [ ] Create new auth package structure
  - [ ] Set up directory layout
  - [ ] Create interface definitions
  - [ ] Add package documentation
- [ ] Database preparation
  - [ ] Create session table migration
  - [ ] Add indexes for performance
  - [ ] Set up cleanup job

### Phase 2: Core Implementation
- [ ] Session Repository
  - [ ] Implement SessionRepository interface
  - [ ] Add CRUD operations
  - [ ] Write unit tests
  - [ ] Write integration tests
- [ ] Session Service
  - [ ] Implement SessionManager interface
  - [ ] Add session lifecycle methods
  - [ ] Write unit tests
  - [ ] Add monitoring hooks
- [ ] Middleware
  - [ ] Implement auth middleware
  - [ ] Add session validation
  - [ ] Write middleware tests

### Phase 3: Integration & Testing
- [ ] End-to-end Testing
  - [ ] Write auth flow tests
  - [ ] Test concurrent sessions
  - [ ] Test session expiry
  - [ ] Test error cases
- [ ] Security Implementation
  - [ ] Set secure cookie options
  - [ ] Add CSRF protection
  - [ ] Implement session cleanup

### Phase 4: Code Cleanup
- [ ] Remove Old Code
  - [ ] Identify all session-related code in `/internal/session/`
  - [ ] List all imports of old session package
  - [ ] Remove old session.go implementation
  - [ ] Update import paths in dependent files
  - [ ] Delete unused session package files
- [ ] Clean Up Tests
  - [ ] Remove old session tests
  - [ ] Clean up any mock implementations
  - [ ] Remove unused test utilities
  - [ ] Update test imports
- [ ] Dependency Cleanup
  - [ ] Remove unused session-related dependencies
  - [ ] Update go.mod and go.sum
  - [ ] Run `go mod tidy`
- [ ] Code Quality
  - [ ] Run linters on new code
  - [ ] Fix any style issues
  - [ ] Ensure consistent naming
  - [ ] Check for dead code

### Phase 5: Documentation
- [ ] Documentation
  - [ ] Update API documentation
  - [ ] Update OpenAPI spec
  - [ ] Add code comments
  - [ ] Update README if needed
