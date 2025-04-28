package session

import (
	"context"
	"time"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/rs/zerolog/log"
)

const (
	// DefaultCleanupInterval is the default interval between cleanup runs
	DefaultCleanupInterval = 15 * time.Minute
)

// CleanupWorker periodically cleans up expired sessions
type CleanupWorker struct {
	store    session.Store
	interval time.Duration
	stop     chan struct{}
}

// NewCleanupWorker creates a new cleanup worker
func NewCleanupWorker(store session.Store, interval time.Duration) *CleanupWorker {
	if interval <= 0 {
		interval = DefaultCleanupInterval
	}

	return &CleanupWorker{
		store:    store,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start starts the cleanup worker
func (s *CleanupWorker) Start(ctx context.Context) {
	log.Info().
		Dur("interval", s.interval).
		Msg("Session cleanup worker started")

	// Run initial cleanup
	if err := s.cleanup(); err != nil {
		log.Error().
			Err(err).
			Msg("Initial session cleanup failed")
	}

	// Start periodic cleanup
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := s.cleanup(); err != nil {
					log.Error().
						Err(err).
						Msg("Session cleanup failed")
				}
			case <-s.stop:
				ticker.Stop()
				log.Info().Msg("Session cleanup worker stopped")
				return
			case <-ctx.Done():
				ticker.Stop()
				log.Info().Msg("Session cleanup worker stopped due to context cancellation")
				return
			}
		}
	}()
}

// Stop stops the cleanup worker
func (s *CleanupWorker) Stop() {
	close(s.stop)
}

// cleanup performs a single cleanup operation
func (s *CleanupWorker) cleanup() error {
	log.Debug().
		Dur("interval", s.interval).
		Msg("Starting session cleanup")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := s.store.Cleanup(ctx)
	if err != nil {
		return err
	}

	log.Debug().
		Dur("interval", s.interval).
		Msg("Session cleanup completed")
	return nil
}
