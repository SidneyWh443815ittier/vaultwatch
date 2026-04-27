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

func containsSubstr(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected %q to contain %q", haystack, needle)
	}
}

func TestGoogleChatSender_PostsFormattedMessage(t *testing.T) {
	var received map[string]string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewGoogleChatSenderWithURL(ts.URL)
	err := s.Send(sender.Alert{
		LeaseID: "secret/db/creds",
		Level:   "critical",
		Message: "lease expiring soon",
		TTL:     30 * time.Minute,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text, ok := received["text"]
	if !ok {
		t.Fatal("expected 'text' field in payload")
	}
	containsSubstr(t, text, "critical")
	containsSubstr(t, text, "secret/db/creds")
	containsSubstr(t, text, "lease expiring soon")
}

func TestGoogleChatSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	s := sender.NewGoogleChatSenderWithURL(ts.URL)
	err := s.Send(sender.Alert{
		LeaseID: "secret/db/creds",
		Level:   "warning",
		Message: "expiring",
		TTL:     time.Hour,
	})

	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
	containsSubstr(t, err.Error(), "403")
}
