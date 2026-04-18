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

func TestOpsGenieSender_PostsCorrectPayload(t *testing.T) {
	var received map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &received)
		if r.Header.Get("Authorization") != "GenieKey test-key" {
			t.Errorf("expected GenieKey auth header")
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	s := sender.NewOpsGenieSenderWithURL("test-key", ts.URL)
	err := s.Send(alert.Alert{
		LeaseID: "secret/my-app/db",
		Level:   alert.LevelCritical,
		TTL:     10 * time.Minute,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["priority"] != "P1" {
		t.Errorf("expected priority P1, got %v", received["priority"])
	}
}

func TestOpsGenieSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	s := sender.NewOpsGenieSenderWithURL("bad-key", ts.URL)
	err := s.Send(alert.Alert{
		LeaseID: "secret/x",
		Level:   alert.LevelWarning,
		TTL:     30 * time.Minute,
	})
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
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &received)
			w.WriteHeader(http.StatusAccepted)
		}))
		s := sender.NewOpsGenieSenderWithURL("k", ts.URL)
		s.Send(alert.Alert{Level: tc.level, TTL: time.Minute})
		if received["priority"] != tc.wantPrio {
			t.Errorf("level %s: want %s got %v", tc.level, tc.wantPrio, received["priority"])
		}
		ts.Close()
	}
}
