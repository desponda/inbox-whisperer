package provider

import (
	"context"
	"errors"

	"github.com/desponda/inbox-whisperer/internal/models"
)

type Type string

const (
	Gmail   Type = "gmail"
	Outlook Type = "outlook"
)

// Capabilities represents what features a provider supports
type Capabilities struct {
	SupportsSearch    bool
	SupportsFolders   bool
	SupportsLabels    bool
	SupportsThreading bool
}

// Config represents a user's linked provider account
// (provider-specific configuration, tokens, etc)
type Config struct {
	UserID string
	Type   Type
	Config map[string]interface{}
}

// Provider is the interface for any email provider implementation
// (Gmail, Outlook, etc)
type Provider interface {
	GetType() Type
	GetCapabilities() Capabilities
	FetchSummaries(ctx context.Context, userID string) ([]models.EmailSummary, error)
	FetchMessage(ctx context.Context, userToken interface{}, messageID string) (*models.EmailMessage, error)
}

// Common errors
var (
	ErrNotFound = errors.New("not found")
)

// ProviderFactoryFunc is a function that returns a Provider given a Config
// Used for registration of provider implementations
type ProviderFactoryFunc func(cfg Config) (Provider, error)

// Factory manages provider registration and user-provider links
// Implements the provider factory pattern
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
