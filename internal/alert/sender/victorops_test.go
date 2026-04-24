package sender_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestVictorOpsSender_PostsCorrectPayload(t *testing.T) {
	var received map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewVictorOpsSenderWithURL(ts.URL)
	a := alert.Alert{
		LeaseID: "secret/prod/api-key",
		TTL:     30 * time.Minute,
		Level:   alert.LevelCritical,
	}

	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["message_type"] != "CRITICAL" {
		t.Errorf("expected CRITICAL, got %v", received["message_type"])
	}
	if received["monitoring_tool"] != "vaultwatch" {
		t.Errorf("expected vaultwatch, got %v", received["monitoring_tool"])
	}
	if received["entity_id"] != "secret/prod/api-key" {
		t.Errorf("unexpected entity_id: %v", received["entity_id"])
	}
}

func TestVictorOpsSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewVictorOpsSenderWithURL(ts.URL)
	err := s.Send(alert.Alert{Level: alert.LevelWarning, TTL: time.Hour})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestVictorOpsSender_MessageTypeMapping(t *testing.T) {
	cases := []struct {
		level   alert.Level
		wantMsg string
	}{
		{alert.LevelCritical, "CRITICAL"},
		{alert.LevelWarning, "WARNING"},
		{alert.LevelOK, "INFO"},
	}

	for _, tc := range cases {
		var received map[string]interface{}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&received)
			w.WriteHeader(http.StatusOK)
		}))

		s := sender.NewVictorOpsSenderWithURL(ts.URL)
		s.Send(alert.Alert{Level: tc.level, TTL: time.Minute})
		ts.Close()

		if received["message_type"] != tc.wantMsg {
			t.Errorf("level %s: want %s, got %v", tc.level, tc.wantMsg, received["message_type"])
		}
	}
}
