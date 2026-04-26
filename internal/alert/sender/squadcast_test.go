package sender_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestSquadcastSender_PostsCorrectPayload(t *testing.T) {
	var received map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewSquadcastSender(ts.URL)
	a := alert.Alert{
		LeaseID:  "auth/token/abc123",
		Level:    alert.LevelCritical,
		TimeLeft: "5m",
		TTL:      "1h",
	}

	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["message"] == "" {
		t.Error("expected non-empty message")
	}
	if received["status"] != "trigger" {
		t.Errorf("expected status=trigger, got %v", received["status"])
	}
	tags, ok := received["tags"].(map[string]interface{})
	if !ok {
		t.Fatal("expected tags to be a map")
	}
	if tags["level"] != string(alert.LevelCritical) {
		t.Errorf("expected tags.level=%s, got %v", alert.LevelCritical, tags["level"])
	}
}

func TestSquadcastSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewSquadcastSender(ts.URL)
	err := s.Send(alert.Alert{
		LeaseID: "secret/db/creds",
		Level:   alert.LevelWarning,
	})
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestSquadcastSender_StatusMapping(t *testing.T) {
	cases := []struct {
		level    alert.Level
		wantStatus string
	}{
		{alert.LevelCritical, "trigger"},
		{alert.LevelWarning, "trigger"},
		{alert.LevelOK, "resolve"},
	}

	for _, tc := range cases {
		var received map[string]interface{}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&received)
			w.WriteHeader(http.StatusOK)
		}))

		s := sender.NewSquadcastSender(ts.URL)
		s.Send(alert.Alert{Level: tc.level, LeaseID: "test/lease"})
		ts.Close()

		if received["status"] != tc.wantStatus {
			t.Errorf("level %s: expected status=%s, got %v", tc.level, tc.wantStatus, received["status"])
		}
	}
}
