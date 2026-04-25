package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
	"github.com/yourusername/vaultwatch/internal/monitor"
)

func TestVictorOpsSender_PostsCorrectPayload(t *testing.T) {
	var received map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := sender.NewVictorOpsSenderWithURL(srv.URL, "team-routing")
	lease := monitor.LeaseInfo{
		LeaseID: "aws/creds/my-role/abc123",
		TTL:     30 * time.Minute,
	}

	if err := s.Send(sender.LevelCritical, lease); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["message_type"] != "CRITICAL" {
		t.Errorf("expected message_type CRITICAL, got %v", received["message_type"])
	}
	if received["entity_id"] != lease.LeaseID {
		t.Errorf("expected entity_id %q, got %v", lease.LeaseID, received["entity_id"])
	}
}

func TestVictorOpsSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := sender.NewVictorOpsSenderWithURL(srv.URL, "team-routing")
	lease := monitor.LeaseInfo{LeaseID: "db/creds/role/xyz", TTL: 10 * time.Minute}

	if err := s.Send(sender.LevelWarning, lease); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestVictorOpsSender_MessageTypeMapping(t *testing.T) {
	cases := []struct {
		level    sender.Level
		wantType string
	}{
		{sender.LevelCritical, "CRITICAL"},
		{sender.LevelWarning, "WARNING"},
		{sender.LevelOK, "INFO"},
	}

	for _, tc := range cases {
		t.Run(tc.wantType, func(t *testing.T) {
			var received map[string]interface{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &received)
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			s := sender.NewVictorOpsSenderWithURL(srv.URL, "key")
			_ = s.Send(tc.level, monitor.LeaseInfo{LeaseID: "id", TTL: time.Minute})

			if received["message_type"] != tc.wantType {
				t.Errorf("level %v: want %q, got %v", tc.level, tc.wantType, received["message_type"])
			}
		})
	}
}
