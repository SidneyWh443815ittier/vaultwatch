package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestGoogleChatSender_PostsFormattedMessage(t *testing.T) {
	var captured []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewGoogleChatSenderWithURL(ts.URL, ts.Client())
	alert := sender.Alert{
		Level:   "critical",
		LeaseID: "secret/db/creds",
		Message: "lease expiring soon",
		TTL:     "5m",
	}

	if err := s.Send(alert); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload sender.GoogleChatPayload
	if err := json.Unmarshal(captured, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !containsSubstr(payload.Text, "critical") {
		t.Errorf("expected level in text, got: %s", payload.Text)
	}
	if !containsSubstr(payload.Text, "secret/db/creds") {
		t.Errorf("expected lease ID in text, got: %s", payload.Text)
	}
	if !containsSubstr(payload.Text, "5m") {
		t.Errorf("expected TTL in text, got: %s", payload.Text)
	}
}

func TestGoogleChatSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	s := sender.NewGoogleChatSenderWithURL(ts.URL, ts.Client())
	err := s.Send(sender.Alert{Level: "warning", LeaseID: "x", Message: "y", TTL: "1m"})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected 403 in error, got: %v", err)
	}
}

func containsSubstr(s, sub string) bool {
	return strings.Contains(s, sub)
}
