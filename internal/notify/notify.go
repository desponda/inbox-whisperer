package notify

import (
	"sync"
)

// In-memory per-user sync status flag (for demo/prototype)
var gmailSyncStatus = struct {
	m sync.RWMutex
	status map[string]bool
}{status: make(map[string]bool)}

// SetGmailSyncStatus sets the sync complete flag for a user
func SetGmailSyncStatus(userID string) {
	gmailSyncStatus.m.Lock()
	defer gmailSyncStatus.m.Unlock()
	gmailSyncStatus.status[userID] = true
}

// ClearGmailSyncStatus clears the sync complete flag for a user
func ClearGmailSyncStatus(userID string) {
	gmailSyncStatus.m.Lock()
	defer gmailSyncStatus.m.Unlock()
	gmailSyncStatus.status[userID] = false
}

// CheckAndClearGmailSyncStatus returns sync status for the user and clears it
func CheckAndClearGmailSyncStatus(userID string) bool {
	gmailSyncStatus.m.Lock()
	defer gmailSyncStatus.m.Unlock()
	status := gmailSyncStatus.status[userID]
	gmailSyncStatus.status[userID] = false
	return status
}
