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

func TestTeamsSender_PostsFormattedMessage(t *testing.T) {
	var got map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewTeamsSender(ts.URL)
	a := alert.Alert{
		LeaseID: "secret/my-app/db",
		Level:   alert.LevelCritical,
		TTL:     5 * time.Minute,
	}
	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["@type"] != "MessageCard" {
		t.Errorf("expected @type MessageCard, got %v", got["@type"])
	}
	if got["themeColor"] != "FF0000" {
		t.Errorf("expected red color for critical, got %v", got["themeColor"])
	}
}

func TestTeamsSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewTeamsSender(ts.URL)
	err := s.Send(alert.Alert{LeaseID: "x", Level: alert.LevelWarning, TTL: time.Minute})
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestTeamsSender_ColorMapping(t *testing.T) {
	cases := []struct {
		level    alert.Level
		expected string
	}{
		{alert.LevelCritical, "FF0000"},
		{alert.LevelWarning, "FFA500"},
		{alert.LevelOK, "00FF00"},
	}
	for _, tc := range cases {
		var got map[string]interface{}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &got)
			w.WriteHeader(http.StatusOK)
		}))
		s := sender.NewTeamsSender(ts.URL)
		_ = s.Send(alert.Alert{LeaseID: "x", Level: tc.level, TTL: time.Minute})
		if got["themeColor"] != tc.expected {
			t.Errorf("level %s: expected color %s, got %v", tc.level, tc.expected, got["themeColor"])
		}
		ts.Close()
	}
}
