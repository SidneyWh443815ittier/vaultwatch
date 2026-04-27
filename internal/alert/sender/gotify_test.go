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

func TestGotifySender_PostsCorrectPayload(t *testing.T) {
	var gotBody map[string]interface{}
	var gotToken string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.URL.Query().Get("token")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := sender.NewGotifySender(srv.URL, "mytoken")
	err := s.Send(alert.Alert{
		LeaseID: "secret/db/creds",
		Level:   alert.LevelCritical,
		TTL:     5 * time.Minute,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gotToken != "mytoken" {
		t.Errorf("expected token 'mytoken', got %q", gotToken)
	}
	if gotBody["title"] == "" {
		t.Error("expected non-empty title")
	}
	priority, ok := gotBody["priority"].(float64)
	if !ok {
		t.Fatal("expected priority field in payload")
	}
	if int(priority) != 10 {
		t.Errorf("expected priority 10 for critical, got %v", priority)
	}
}

func TestGotifySender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	s := sender.NewGotifySender(srv.URL, "badtoken")
	err := s.Send(alert.Alert{
		LeaseID: "secret/app/key",
		Level:   alert.LevelWarning,
		TTL:     30 * time.Minute,
	})

	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestGotifySender_PriorityMapping(t *testing.T) {
	cases := []struct {
		level    alert.Level
		wantPrio float64
	}{
		{alert.LevelCritical, 10},
		{alert.LevelWarning, 5},
		{alert.LevelOK, 1},
	}

	for _, tc := range cases {
		var gotBody map[string]interface{}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			w.WriteHeader(http.StatusOK)
		}))

		s := sender.NewGotifySender(srv.URL, "tok")
		_ = s.Send(alert.Alert{Level: tc.level, TTL: time.Minute})
		srv.Close()

		prio, _ := gotBody["priority"].(float64)
		if prio != tc.wantPrio {
			t.Errorf("level %v: expected priority %.0f, got %.0f", tc.level, tc.wantPrio, prio)
		}
	}
}
