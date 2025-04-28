package gmail

import (
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
)

// UsersMessagesGetCall abstracts the Do method for UsersMessagesGet
type UsersMessagesGetCall interface {
	Do(...googleapi.CallOption) (*gmail.Message, error)
}

// UsersMessagesListCall abstracts the Do method for UsersMessagesList
type UsersMessagesListCall interface {
	Do(...googleapi.CallOption) (*gmail.ListMessagesResponse, error)
}

// GmailAPI defines the subset of the Google Gmail API used by GmailService.
type GmailAPI interface {
	UsersMessagesGet(userID, msgID string) UsersMessagesGetCall
	UsersMessagesList(userID string) UsersMessagesListCall
}
