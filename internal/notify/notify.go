package notify

import (
	"sync"
)

var gmailSyncStatus = struct {
	m      sync.RWMutex
	status map[string]bool
}{status: make(map[string]bool)}

func SetGmailSyncStatus(userID string) {
	gmailSyncStatus.m.Lock()
	defer gmailSyncStatus.m.Unlock()
	gmailSyncStatus.status[userID] = true
}

func ClearGmailSyncStatus(userID string) {
	gmailSyncStatus.m.Lock()
	defer gmailSyncStatus.m.Unlock()
	gmailSyncStatus.status[userID] = false
}

func CheckAndClearGmailSyncStatus(userID string) bool {
	gmailSyncStatus.m.Lock()
	defer gmailSyncStatus.m.Unlock()
	status := gmailSyncStatus.status[userID]
	gmailSyncStatus.status[userID] = false
	return status
}
