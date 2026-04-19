package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestRocketChatSender_PostsFormattedMessage(t *testing.T) {
	var got map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewRocketChatSender(ts.URL)
	a := alert.Alert{
		LeaseID: "secret/my-app/db",
		Level:   alert.Critical,
		TTL:     30 * time.Minute,
	}

	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got["text"] == nil {
		t.Fatal("expected text field in payload")
	}
	attachments, ok := got["attachments"].([]interface{})
	if !ok || len(attachments) == 0 {
		t.Fatal("expected at least one attachment")
	}
	attach := attachments[0].(map[string]interface{})
	if attach["color"] != "#d00000" {
		t.Errorf("expected critical color, got %v", attach["color"])
	}
}

func TestRocketChatSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewRocketChatSender(ts.URL)
	a := alert.Alert{
		LeaseID: "secret/my-app/db",
		Level:   alert.Warning,
		TTL:     1 * time.Hour,
	}

	if err := s.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
