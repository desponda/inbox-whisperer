package service

import (
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
)

type GmailService struct{}

func NewGmailService() *GmailService {
	return &GmailService{}
}

// MessageSummary contains minimal info about a Gmail message
// (expand as needed)
type MessageSummary struct {
	ID      string `json:"id"`
	ThreadID string `json:"thread_id"`
	Snippet string `json:"snippet"`
	From    string `json:"from"`
	Subject string `json:"subject"`
}

// FetchMessages fetches the latest 10 messages using the user's OAuth2 token
func (s *GmailService) FetchMessages(ctx context.Context, token *oauth2.Token) ([]MessageSummary, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://gmail.googleapis.com/gmail/v1/users/me/messages?maxResults=10")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var msgList struct {
		Messages []struct {
			ID string `json:"id"`
			ThreadID string `json:"threadId"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&msgList); err != nil {
		return nil, err
	}
	var summaries []MessageSummary
	for _, m := range msgList.Messages {
		msgResp, err := client.Get("https://gmail.googleapis.com/gmail/v1/users/me/messages/" + m.ID + "?format=metadata&metadataHeaders=Subject&metadataHeaders=From&metadataHeaders=To")
		if err != nil {
			continue
		}
		var msg struct {
			ID      string `json:"id"`
			ThreadID string `json:"threadId"`
			Snippet string `json:"snippet"`
			Payload struct {
				Headers []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"headers"`
			} `json:"payload"`
		}
		if err := json.NewDecoder(msgResp.Body).Decode(&msg); err != nil {
			msgResp.Body.Close()
			continue
		}
		msgResp.Body.Close()
		var from, subject string
		for _, h := range msg.Payload.Headers {
			if h.Name == "From" {
				from = h.Value
			} else if h.Name == "Subject" {
				subject = h.Value
			}
		}
		summaries = append(summaries, MessageSummary{
			ID:      msg.ID,
			ThreadID: msg.ThreadID,
			Snippet: msg.Snippet,
			From:    from,
			Subject: subject,
		})
	}
	return summaries, nil
}
