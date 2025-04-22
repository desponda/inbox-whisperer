package session

import (
	"net/http"
	"net/http/cookiejar"
)

// NewTestCookieJar returns a new cookie jar for test clients.
func NewTestCookieJar() (http.CookieJar, error) {
	return cookiejar.New(nil)
}
