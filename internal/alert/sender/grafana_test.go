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

func TestGrafanaSender_PostsCorrectPayload(t *testing.T) {
	var got map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing or wrong Authorization header")
		}
		json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewGrafanaSenderWithURL(ts.URL, "test-key")
	err := s.Send(alert.Notification{
		LeaseID: "secret/my-token",
		TTL:     2 * time.Hour,
		Level:   alert.LevelCritical,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["state"] != "alerting" {
		t.Errorf("expected state=alerting, got %s", got["state"])
	}
	if got["title"] == "" {
		t.Error("expected non-empty title")
	}
}

func TestGrafanaSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewGrafanaSenderWithURL(ts.URL, "key")
	err := s.Send(alert.Notification{Level: alert.LevelWarning, TTL: time.Minute})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestGrafanaSender_StateMapping(t *testing.T) {
	cases := []struct {
		level    alert.Level
		wantState string
	}{
		{alert.LevelCritical, "alerting"},
		{alert.LevelWarning, "pending"},
		{alert.LevelOK, "ok"},
	}
	for _, tc := range cases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var p map[string]string
			json.NewDecoder(r.Body).Decode(&p)
			if p["state"] != tc.wantState {
				t.Errorf("level %v: want state %s, got %s", tc.level, tc.wantState, p["state"])
			}
			w.WriteHeader(http.StatusOK)
		}))
		s := sender.NewGrafanaSenderWithURL(ts.URL, "k")
		s.Send(alert.Notification{Level: tc.level, TTL: time.Minute})
		ts.Close()
	}
}
