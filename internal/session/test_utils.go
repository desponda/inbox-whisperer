package session

// TestStore returns a pointer to the in-memory session store for use in integration tests.
import "sync"

// TestStore returns a pointer to the in-memory session store for use in integration tests.
func TestStore() *struct {
	sync.RWMutex
	data map[string]*SessionData
} {
	return &store
}
