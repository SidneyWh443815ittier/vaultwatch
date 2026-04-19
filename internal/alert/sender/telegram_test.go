package sender_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func newTelegramSenderWithBase(token, chatID, base string) alert.Sender {
	return sender.NewTelegramSenderWithBase(token, chatID, base)
}

func TestTelegramSender_PostsFormattedMessage(t *testing.T) {
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := sender.NewTelegramSenderWithBase("mytoken", "chat123", srv.URL)
	err := s.Send(alert.Alert{
		Level:   "critical",
		LeaseID: "secret/my-secret",
		TTL:     "2h",
		Message: "expires soon",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["chat_id"] != "chat123" {
		t.Errorf("expected chat_id=chat123, got %q", gotBody["chat_id"])
	}
	if !strings.Contains(gotBody["text"], "secret/my-secret") {
		t.Errorf("expected text to contain lease ID, got %q", gotBody["text"])
	}
	if gotBody["parse_mode"] != "Markdown" {
		t.Errorf("expected parse_mode=Markdown, got %q", gotBody["parse_mode"])
	}
}

func TestTelegramSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	s := sender.NewTelegramSenderWithBase("badtoken", "chat123", srv.URL)
	err := s.Send(alert.Alert{Level: "warning", LeaseID: "lease/x", TTL: "1h", Message: "warn"})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
