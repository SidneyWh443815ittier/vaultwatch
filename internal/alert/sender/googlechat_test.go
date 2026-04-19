package sender_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestGoogleChatSender_PostsFormattedMessage(t *testing.T) {
	var received map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewGoogleChatSender(ts.URL)
	err := s.Send(alert.Alert{
		Level:   alert.LevelCritical,
		LeaseID: "secret/db",
		TTL:     "5m",
		Message: "lease expiring soon",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["text"] == "" {
		t.Error("expected non-empty text field")
	}
}

func TestGoogleChatSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewGoogleChatSender(ts.URL)
	err := s.Send(alert.Alert{
		Level:   alert.LevelWarning,
		LeaseID: "secret/token",
		TTL:     "30m",
		Message: "warning",
	})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
