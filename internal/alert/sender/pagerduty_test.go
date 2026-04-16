package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestPagerDutySender_PostsCorrectPayload(t *testing.T) {
	var received map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	s := sender.NewPagerDutySenderWithURL("test-key", ts.URL)
	err := s.Send("critical", "vault lease expiring soon")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if received["routing_key"] != "test-key" {
		t.Errorf("expected routing_key 'test-key', got %v", received["routing_key"])
	}
	if received["event_action"] != "trigger" {
		t.Errorf("expected event_action 'trigger', got %v", received["event_action"])
	}
	payload := received["payload"].(map[string]interface{})
	if payload["severity"] != "critical" {
		t.Errorf("expected severity 'critical', got %v", payload["severity"])
	}
	if payload["summary"] != "vault lease expiring soon" {
		t.Errorf("unexpected summary: %v", payload["summary"])
	}
}

func TestPagerDutySender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	s := sender.NewPagerDutySenderWithURL("key", ts.URL)
	err := s.Send("warning", "test")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestPagerDutySender_SeverityMapping(t *testing.T) {
	cases := []struct{ level, want string }{
		{"critical", "critical"},
		{"warning", "warning"},
		{"info", "info"},
		{"unknown", "info"},
	}
	for _, c := range cases {
		var received map[string]interface{}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &received)
			w.WriteHeader(http.StatusAccepted)
		}))
		s := sender.NewPagerDutySenderWithURL("k", ts.URL)
		s.Send(c.level, "msg")
		payload := received["payload"].(map[string]interface{})
		if payload["severity"] != c.want {
			t.Errorf("level %q: expected severity %q, got %v", c.level, c.want, payload["severity"])
		}
		ts.Close()
	}
}
