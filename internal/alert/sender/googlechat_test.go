package sender_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestGoogleChatSender_PostsFormattedMessage(t *testing.T) {
	var received map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := sender.NewGoogleChatSenderWithURL(srv.URL)
	s.SetClient(&http.Client{Timeout: 5 * time.Second})

	alert := sender.Alert{
		LeaseID: "secret/myapp/db#abc123",
		Level:   "critical",
		Message: "lease expiring soon",
		TTL:     "5m",
	}

	if err := s.Send(alert); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text, ok := received["text"]
	if !ok {
		t.Fatal("expected 'text' field in payload")
	}
	for _, want := range []string{"critical", "secret/myapp/db#abc123", "lease expiring soon", "5m"} {
		if !containsSubstr(text, want) {
			t.Errorf("text %q missing expected substring %q", text, want)
		}
	}
}

func TestGoogleChatSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := sender.NewGoogleChatSenderWithURL(srv.URL)
	s.SetClient(&http.Client{Timeout: 5 * time.Second})

	err := s.Send(sender.Alert{LeaseID: "x", Level: "warning", Message: "msg", TTL: "1h"})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
