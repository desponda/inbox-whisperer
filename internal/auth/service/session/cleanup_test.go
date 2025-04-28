package session

import (
	"context"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/api/testutils"
)

func TestNewCleanupWorker(t *testing.T) {
	store := testutils.NewMockStore()

	tests := []struct {
		name     string
		interval time.Duration
		want     time.Duration
	}{
		{
			name:     "zero interval uses default",
			interval: 0,
			want:     DefaultCleanupInterval,
		},
		{
			name:     "negative interval uses default",
			interval: -1 * time.Second,
			want:     DefaultCleanupInterval,
		},
		{
			name:     "custom interval is respected",
			interval: 5 * time.Minute,
			want:     5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker := NewCleanupWorker(store, tt.interval)
			if worker.interval != tt.want {
				t.Errorf("NewCleanupWorker() interval = %v, want %v", worker.interval, tt.want)
			}
		})
	}
}

func TestCleanupWorker_Start_Stop(t *testing.T) {
	store := testutils.NewMockStore()
	interval := 50 * time.Millisecond
	worker := NewCleanupWorker(store, interval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the worker
	worker.Start(ctx)

	// Wait for at least one cleanup
	time.Sleep(interval + 10*time.Millisecond)

	// Stop the worker
	worker.Stop()

	// Check that cleanup was called at least once
	if calls := store.GetCleanupCalls(); calls < 1 {
		t.Errorf("cleanup was not called, want at least 1 call, got %d", calls)
	}
}

func TestCleanupWorker_ContextCancellation(t *testing.T) {
	store := testutils.NewMockStore()
	interval := 50 * time.Millisecond
	worker := NewCleanupWorker(store, interval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the worker
	worker.Start(ctx)

	// Wait for at least one cleanup
	time.Sleep(interval + 10*time.Millisecond)

	// Cancel context
	cancel()

	// Wait for worker to stop
	time.Sleep(interval)

	// Get the number of calls before stopping
	callsBeforeStop := store.GetCleanupCalls()

	// Wait another interval
	time.Sleep(interval + 10*time.Millisecond)

	// Check that no more cleanups occurred after context cancellation
	if calls := store.GetCleanupCalls(); calls > callsBeforeStop {
		t.Errorf("cleanup was called after context cancellation")
	}
}
