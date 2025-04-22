package session

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionMiddlewareAndSetSession(t *testing.T) {
	var resp3, resp4 *http.Response
	var body2 []byte
	var got2 string

	// Standard session set/get
	mux := http.NewServeMux()
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		SetSession(w, r, "user123", "tokenABC")
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		uid := GetUserID(r.Context())
		tok := GetToken(r.Context())
		w.Write([]byte(uid + ":" + tok))
	})

	ts := httptest.NewServer(Middleware(mux))
	defer ts.Close()

	jar, err := NewTestCookieJar()
	if err != nil {
		t.Fatalf("failed to create cookie jar: %v", err)
	}
	client := &http.Client{Jar: jar}

	resp, err := client.Get(ts.URL + "/set")
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	defer resp.Body.Close()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("no cookies set")
	}
	cookie := cookies[0]
	if cookie.Name != "session_id" || cookie.Value == "" {
		t.Errorf("session_id cookie not set properly")
	}

	// Get session
	resp2, err := client.Get(ts.URL + "/get")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp2.Body.Close()
	body, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	got := string(body)
	if got != "user123:tokenABC" {
		t.Errorf("unexpected session data: %s", got)
	}

	// Test SetSessionValue and GetSessionValue
	mux.HandleFunc("/setval", func(w http.ResponseWriter, r *http.Request) {
		SetSessionValue(w, r, "foo", "bar")
		SetSessionValue(w, r, "baz", "qux")
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/getval", func(w http.ResponseWriter, r *http.Request) {
		foo := GetSessionValue(r, "foo")
		baz := GetSessionValue(r, "baz")
		w.Write([]byte(foo + ":" + baz))
	})

	resp3 = nil
	err = nil
	resp3, err = client.Get(ts.URL + "/setval")
	if err != nil {
		t.Fatalf("setval failed: %v", err)
	}
	defer resp3.Body.Close()

	resp4 = nil
	err = nil
	resp4, err = client.Get(ts.URL + "/getval")
	if err != nil {
		t.Fatalf("getval failed: %v", err)
	}
	defer resp4.Body.Close()
	body2 = nil
	err = nil
	body2, err = io.ReadAll(resp4.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	got2 = string(body2)
	if got2 != "bar:qux" {
		t.Errorf("unexpected session value data: %s", got2)
	}

	// Overwrite session
	resp3 = nil
	err = nil
	resp3, err = client.Get(ts.URL + "/set")
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	defer resp3.Body.Close()

	resp4 = nil
	err = nil
	resp4, err = client.Get(ts.URL + "/get")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp4.Body.Close()
	body2 = nil
	err = nil
	body2, err = io.ReadAll(resp4.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	got2 = string(body2)
	if got2 != "user123:tokenABC" {
		t.Errorf("unexpected session data after overwrite: %s", got2)
	}

	// Invalid/no session cookie
	client2 := &http.Client{} // no jar
	resp5, err := client2.Get(ts.URL + "/get")
	if err != nil {
		t.Fatalf("get with no cookie failed: %v", err)
	}
	defer resp5.Body.Close()
	body3, err := io.ReadAll(resp5.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	got3 := string(body3)
	if got3 != ":" {
		t.Errorf("expected empty session for no cookie, got: %s", got3)
	}

	// Concurrency: simulate multiple sessions
	concurrent := 10
	errs := make(chan error, concurrent)
	for i := 0; i < concurrent; i++ {
		go func(idx int) {
			c := &http.Client{Jar: nil}
			resp, err := c.Get(ts.URL + "/set")
			if err != nil {
				errs <- err
				return
			}
			defer resp.Body.Close()
			resp, err = c.Get(ts.URL + "/get")
			if err != nil {
				errs <- err
				return
			}
			defer resp.Body.Close()
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				errs <- err
				return
			}
			if string(b) != "user123:tokenABC" && string(b) != ":" {
				errs <- io.ErrUnexpectedEOF
				return
			}
			errs <- nil
		}(i)
	}
	for i := 0; i < concurrent; i++ {
		err := <-errs
		if err != nil {
			t.Errorf("concurrent session error: %v", err)
		}
	}
}
