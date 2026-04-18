package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/vaultwatch/internal/alert/sender"
)

func TestVictorOpsSender_PostsCorrectPayload(t *testing.T) {
	var captured map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewVictorOpsSenderWithURL(ts.URL, "default")
	if err := s.Send("critical", "lease expires soon"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured["message_type"] != "CRITICAL" {
		t.Errorf("expected CRITICAL, got %s", captured["message_type"])
	}
	if captured["state_message"] != "lease expires soon" {
		t.Errorf("unexpected state_message: %s", captured["state_message"])
	}
}

func TestVictorOpsSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewVictorOpsSenderWithURL(ts.URL, "default")
	if err := s.Send("warning", "test"); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestVictorOpsSender_MessageTypeMapping(t *testing.T) {
	cases := []struct {
		level    string
		expected string
	}{
		{"critical", "CRITICAL"},
		{"warning", "WARNING"},
		{"info", "INFO"},
		{"unknown", "INFO"},
	}
	for _, tc := range cases {
		var captured map[string]string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &captured)
			w.WriteHeader(http.StatusOK)
		}))
		s := sender.NewVictorOpsSenderWithURL(ts.URL, "default")
		_ = s.Send(tc.level, "msg")
		ts.Close()
		if captured["message_type"] != tc.expected {
			t.Errorf("level %s: expected %s, got %s", tc.level, tc.expected, captured["message_type"])
		}
	}
}
