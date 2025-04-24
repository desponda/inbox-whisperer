package service

import (
	"context"
	"errors"
	"sync"

	"github.com/desponda/inbox-whisperer/internal/service/gmail"
)

type EmailProvider = gmail.EmailProvider

type ProviderType string

const (
	ProviderGmail   ProviderType = "gmail"
	ProviderOutlook ProviderType = "outlook"
)

// ProviderConfig represents a user's linked provider account (simplified)
type ProviderConfig struct {
	UserID string
	Type   ProviderType
	// ...tokens, config, etc.
}

// EmailProviderFactory returns providers for a user
// (in real usage, would query DB for user's linked accounts)
type EmailProviderFactory struct {
	mu       sync.RWMutex
	creators map[ProviderType]func(cfg ProviderConfig) (EmailProvider, error)
	// In-memory mapping for demo; replace with DB in prod
	linked map[string][]ProviderConfig // userID -> []ProviderConfig
}

func NewEmailProviderFactory() *EmailProviderFactory {
	return &EmailProviderFactory{
		creators: make(map[ProviderType]func(cfg ProviderConfig) (EmailProvider, error)),
		linked:   make(map[string][]ProviderConfig),
	}
}

// RegisterProvider allows registration of a provider constructor
func (f *EmailProviderFactory) RegisterProvider(ptype ProviderType, creator func(cfg ProviderConfig) (EmailProvider, error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.creators[ptype] = creator
}

// LinkProvider links a provider to a user (demo, not persistent)
func (f *EmailProviderFactory) LinkProvider(userID string, cfg ProviderConfig) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.linked[userID] = append(f.linked[userID], cfg)
}

func (f *EmailProviderFactory) ProvidersForUser(ctx context.Context, userID string) ([]EmailProvider, error) {
	f.mu.RLock()
	linked := f.linked[userID]
	f.mu.RUnlock()
	if len(linked) == 0 {
		return nil, errors.New("no providers linked for user")
	}
	providers := make([]EmailProvider, 0, len(linked))
	for _, cfg := range linked {
		f.mu.RLock()
		creator, ok := f.creators[cfg.Type]
		f.mu.RUnlock()
		if !ok {
			continue // skip unknown provider
		}
		prov, err := creator(cfg)
		if err != nil {
			continue // skip on error
		}
		providers = append(providers, prov)
	}
	if len(providers) == 0 {
		return nil, errors.New("no valid providers for user")
	}
	return providers, nil
}
