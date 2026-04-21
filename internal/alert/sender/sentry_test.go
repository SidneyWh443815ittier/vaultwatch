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

func TestSentrySender_PostsCorrectPayload(t *testing.T) {
	var received map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing or wrong Authorization header")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewSentrySenderWithURL("test-token", "my-org", "my-project", ts.URL)

	msg := alert.Message{
		LeaseID: "lease/abc/123",
		Level:   alert.LevelCritical,
		TTL:     5 * time.Minute,
		Body:    "lease expiring soon",
	}

	if err := s.Send(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["title"] != "lease/abc/123" {
		t.Errorf("expected title lease/abc/123, got %v", received["title"])
	}
	if received["level"] != "fatal" {
		t.Errorf("expected level fatal, got %v", received["level"])
	}
	if received["platform"] != "other" {
		t.Errorf("expected platform other, got %v", received["platform"])
	}
}

func TestSentrySender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewSentrySenderWithURL("tok", "org", "proj", ts.URL)
	msg := alert.Message{LeaseID: "x", Level: alert.LevelWarning, TTL: time.Minute}

	if err := s.Send(msg); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestSentrySender_LevelMapping(t *testing.T) {
	cases := []struct {
		level    alert.Level
		wantStr  string
	}{
		{alert.LevelCritical, "fatal"},
		{alert.LevelWarning, "warning"},
		{alert.LevelOK, "info"},
	}

	for _, tc := range cases {
		t.Run(tc.wantStr, func(t *testing.T) {
			var received map[string]interface{}
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&received)
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			s := sender.NewSentrySenderWithURL("tok", "org", "proj", ts.URL)
			_ = s.Send(alert.Message{LeaseID: "id", Level: tc.level, TTL: time.Minute})

			if received["level"] != tc.wantStr {
				t.Errorf("expected %q, got %v", tc.wantStr, received["level"])
			}
		})
	}
}
