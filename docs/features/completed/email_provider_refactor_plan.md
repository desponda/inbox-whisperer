# Email Provider Refactor & Restoration Plan

## Objective
Restore and future-proof the backend email service architecture for Inbox Whisperer by introducing a provider-agnostic interface, factory pattern, summary/detail endpoints, and robust caching. This will enable multi-provider support (Gmail, Outlook, etc.), efficient email listing, and scalable extensibility.

---

## 1. Provider Interface & Factory Design

### EmailProvider Interface
Defines the contract for any email provider implementation (Gmail, Outlook, etc.):

```go
// EmailProvider abstracts any email provider (Gmail, Outlook, etc.)
type EmailProvider interface {
    // FetchSummaries fetches summary info for a user's emails (minimal fields, paginated)
    FetchSummaries(ctx context.Context, userID string, params FetchParams) ([]EmailSummary, error)

    // FetchMessage fetches the full content of a specific email by ID
    FetchMessage(ctx context.Context, userID, messageID string) (EmailMessage, error)
}

// FetchParams can include pagination, filters, etc.
type FetchParams struct {
    AfterID           string
    AfterInternalDate int64
    Limit             int
}

// EmailSummary is a minimal DTO for list endpoints
// (subject, sender, snippet, date, etc.)
type EmailSummary struct {
    ID            string
    ThreadID      string
    Snippet       string
    From          string
    Subject       string
    InternalDate  int64
    Provider      string // e.g. "gmail", "outlook"
}

// EmailMessage is the full email content DTO
// (all fields, body, attachments, etc.)
type EmailMessage struct {
    ID            string
    ThreadID      string
    From          string
    To            string
    Subject       string
    Date          string
    Body          string
    Snippet       string
    InternalDate  int64
    Provider      string
    // ...attachments, labels, etc.
}
```

### Provider Factory
Responsible for selecting/instantiating the correct provider(s) for a user/account at runtime:

```go
// ProviderFactoryFunc is a function that returns a Provider given a Config
type ProviderFactoryFunc func(cfg Config) (Provider, error)

// Factory manages provider registration and user-provider links
type Factory struct {
    providers map[Type]ProviderFactoryFunc
    userLinks map[string][]Config // userID -> []Config
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *Factory {
    return &Factory{
        providers: make(map[Type]ProviderFactoryFunc),
        userLinks: make(map[string][]Config),
    }
}

// RegisterProvider registers a provider factory for a given type
func (f *Factory) RegisterProvider(t Type, factory ProviderFactoryFunc) {
    f.providers[t] = factory
}

// LinkProvider links a user to a provider config
func (f *Factory) LinkProvider(userID string, cfg Config) {
    f.userLinks[userID] = append(f.userLinks[userID], cfg)
}

// ProviderForUser returns a Provider for a user and provider type
func (f *Factory) ProviderForUser(ctx context.Context, userID string, t Type) (Provider, error) {
    cfgs, ok := f.userLinks[userID]
    if !ok {
        return nil, ErrNotFound
    }
    for _, cfg := range cfgs {
        if cfg.Type == t {
            factory, ok := f.providers[t]
            if !ok {
                return nil, errors.New("provider factory not registered")
            }
            return factory(cfg)
        }
    }
    return nil, ErrNotFound
}

// UserLinks returns the linked provider configs for a user
func (f *Factory) UserLinks(userID string) []Config {
    return f.userLinks[userID]
}
```

- Providers are registered at startup (Gmail, Outlook, etc.)
- Factory reads user/provider mapping from DB or config
- Service layer uses factory to fetch from all providers as needed

---

# Email Provider Refactor Plan

**Status: COMPLETE — All objectives in this feature plan have been implemented and tested.**

---

## Done
- [x] Implement the `EmailProvider` interface for Gmail (refactored logic in `gmail_provider.go`)
- [x] Implement provider factory and user/provider mapping (`email_provider_factory.go`)
- [x] Refactor service layer to use provider interface/factory (`email_service_factory.go`)
- [x] Add MultiProviderEmailService to aggregate across all user providers
- [x] Ensure `Date` field is present and populated in all email models
- [x] Wire up provider factory and MultiProviderEmailService in main/server startup
- [x] Register GmailProvider in the factory; stub other providers
- [x] Update API handlers to use new service and context conventions (ensure `user_id` is in context)
- [x] Unify EmailProvider interface and provider FetchMessage signatures (token as interface{})
- [x] Service now supports multi-provider aggregation with unified provider interface
- [x] Implement summary endpoints and use summary DTOs for list responses
- [x] Aggregate and deduplicate results from all providers for a user
- [x] Add/configure summary and detail caching
- [x] Update/add tests and documentation

### Details
- The backend summary endpoint aggregates, deduplicates, and sorts summaries from all linked providers for a user. Summary and detail caching are now implemented for optimal performance and efficiency.

---

*This plan is preserved for historical reference. All items are complete and the feature is live.*
