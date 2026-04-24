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

func TestOpsGenieSender_PostsCorrectPayload(t *testing.T) {
	var received map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if r.Header.Get("Authorization") != "GenieKey test-key" {
			t.Errorf("missing or wrong Authorization header: %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	s := sender.NewOpsGenieSenderWithURL("test-key", ts.URL)
	a := alert.Alert{
		LeaseID: "secret/my-app/db",
		TTL:     2 * time.Hour,
		Level:   alert.LevelCritical,
	}

	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["priority"] != "P1" {
		t.Errorf("expected priority P1, got %v", received["priority"])
	}
	if received["message"] == "" {
		t.Error("expected non-empty message")
	}
}

func TestOpsGenieSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	s := sender.NewOpsGenieSenderWithURL("bad-key", ts.URL)
	err := s.Send(alert.Alert{Level: alert.LevelWarning, TTL: time.Hour})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestOpsGenieSender_PriorityMapping(t *testing.T) {
	cases := []struct {
		level    alert.Level
		wantPrio string
	}{
		{alert.LevelCritical, "P1"},
		{alert.LevelWarning, "P3"},
		{alert.LevelOK, "P5"},
	}

	for _, tc := range cases {
		var received map[string]interface{}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&received)
			w.WriteHeader(http.StatusAccepted)
		}))

		s := sender.NewOpsGenieSenderWithURL("k", ts.URL)
		s.Send(alert.Alert{Level: tc.level, TTL: time.Minute})
		ts.Close()

		if received["priority"] != tc.wantPrio {
			t.Errorf("level %s: want priority %s, got %v", tc.level, tc.wantPrio, received["priority"])
		}
	}
}
