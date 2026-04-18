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

func TestDatadogSender_PostsCorrectPayload(t *testing.T) {
	var received map[string]interface{}
	var apiKey string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey = r.Header.Get("DD-API-KEY")
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	s := sender.NewDatadogSenderWithURL("test-key", ts.URL)
	err := s.Send(alert.Alert{
		LeaseID: "secret/my-app/db",
		TTL:     2 * time.Hour,
		Level:   alert.LevelWarning,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "test-key" {
		t.Errorf("expected api key 'test-key', got %q", apiKey)
	}
	if received["alert_type"] != "warning" {
		t.Errorf("expected alert_type 'warning', got %v", received["alert_type"])
	}
}

func TestDatadogSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	s := sender.NewDatadogSenderWithURL("bad-key", ts.URL)
	err := s.Send(alert.Alert{LeaseID: "x", Level: alert.LevelCritical})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestDatadogSender_AlertTypeMapping(t *testing.T) {
	cases := []struct {
		level    alert.Level
		expected string
	}{
		{alert.LevelCritical, "error"},
		{alert.LevelWarning, "warning"},
		{alert.LevelOK, "info"},
	}

	for _, tc := range cases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if body["alert_type"] != tc.expected {
				t.Errorf("level %v: expected %q, got %v", tc.level, tc.expected, body["alert_type"])
			}
			w.WriteHeader(http.StatusAccepted)
		}))
		s := sender.NewDatadogSenderWithURL("k", ts.URL)
		s.Send(alert.Alert{Level: tc.level})
		ts.Close()
	}
}
