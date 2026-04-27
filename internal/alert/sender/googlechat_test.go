package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func containsSubstr(s, sub string) bool {
	return strings.Contains(s, sub)
}

func TestGoogleChatSender_PostsFormattedMessage(t *testing.T) {
	var received map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := sender.NewGoogleChatSenderWithURL(srv.URL)
	err := s.Send(sender.Alert{
		LeaseID: "secret/data/myapp",
		Level:   "warning",
		Message: "Lease expiring soon",
		TTL:     2 * time.Hour,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text, ok := received["text"]
	if !ok {
		t.Fatal("expected 'text' field in payload")
	}
	if !containsSubstr(text, "warning") {
		t.Errorf("expected level in text, got: %s", text)
	}
	if !containsSubstr(text, "secret/data/myapp") {
		t.Errorf("expected lease ID in text, got: %s", text)
	}
}

func TestGoogleChatSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	s := sender.NewGoogleChatSenderWithURL(srv.URL)
	err := s.Send(sender.Alert{
		LeaseID: "secret/data/myapp",
		Level:   "critical",
		Message: "Lease expired",
		TTL:     0,
	})

	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
	if !containsSubstr(err.Error(), "403") {
		t.Errorf("expected status 403 in error, got: %v", err)
	}
}
