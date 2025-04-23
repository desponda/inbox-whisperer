package api

import (
	"net/http"
	"github.com/desponda/inbox-whisperer/internal/notify"
)

// GmailSyncStatusHandler returns sync status for the current user and clears it
func GmailSyncStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextUserIDKey).(string)
	if notify.CheckAndClearGmailSyncStatus(userID) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("sync_complete"))
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
