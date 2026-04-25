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
	"github.com/yourusername/vaultwatch/internal/monitor"
)

func TestMattermostSender_PostsFormattedMessage(t *testing.T) {
	var received map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := sender.NewMattermostSender(srv.URL)
	lease := monitor.LeaseInfo{
		LeaseID: "secret/data/myapp/db",
		TTL:     45 * time.Minute,
	}

	if err := s.Send(sender.LevelWarning, lease); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text, ok := received["text"].(string)
	if !ok || text == "" {
		t.Fatalf("expected non-empty text field, got %v", received["text"])
	}
	if !strings.Contains(text, lease.LeaseID) {
		t.Errorf("text %q does not contain lease ID %q", text, lease.LeaseID)
	}
	if received["username"] != "vaultwatch" {
		t.Errorf("expected username vaultwatch, got %v", received["username"])
	}
}

func TestMattermostSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	s := sender.NewMattermostSender(srv.URL)
	lease := monitor.LeaseInfo{LeaseID: "kv/my-secret", TTL: 5 * time.Minute}

	if err := s.Send(sender.LevelCritical, lease); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
