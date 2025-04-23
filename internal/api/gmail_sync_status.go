package api

import (
	"net/http"
	"github.com/desponda/inbox-whisperer/internal/notify"
)

// GmailSyncStatusHandler returns sync status for the current user and clears it
func GmailSyncStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := ValidateAuth(r)
	if err != nil {
		RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if notify.CheckAndClearGmailSyncStatus(userID) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("sync_complete"))
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
