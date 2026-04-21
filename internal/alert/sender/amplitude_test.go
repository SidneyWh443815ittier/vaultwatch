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

func TestAmplitudeSender_PostsCorrectPayload(t *testing.T) {
	var received map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := sender.NewAmplitudeSenderWithURL("test-api-key", srv.URL)
	err := s.Send(alert.Message{
		Level:     "critical",
		LeaseID:   "lease/abc/123",
		ExpiresIn: 5 * time.Minute,
		Summary:   "Lease expiring soon",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if received["api_key"] != "test-api-key" {
		t.Errorf("expected api_key 'test-api-key', got %v", received["api_key"])
	}
	events, ok := received["events"].([]interface{})
	if !ok || len(events) != 1 {
		t.Fatalf("expected 1 event, got %v", received["events"])
	}
	event := events[0].(map[string]interface{})
	if event["event_type"] != "vault_lease_alert" {
		t.Errorf("expected event_type 'vault_lease_alert', got %v", event["event_type"])
	}
	props, _ := event["event_properties"].(map[string]interface{})
	if props["level"] != "critical" {
		t.Errorf("expected level 'critical', got %v", props["level"])
	}
	if props["lease_id"] != "lease/abc/123" {
		t.Errorf("expected lease_id 'lease/abc/123', got %v", props["lease_id"])
	}
}

func TestAmplitudeSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := sender.NewAmplitudeSenderWithURL("key", srv.URL)
	err := s.Send(alert.Message{
		Level:     "warning",
		LeaseID:   "lease/xyz",
		ExpiresIn: 10 * time.Minute,
		Summary:   "Lease warning",
	})

	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestAmplitudeSender_MissingAPIKeyReturnsError(t *testing.T) {
	_, err := sender.NewAmplitudeSender("")
	if err == nil {
		t.Fatal("expected error for missing api_key, got nil")
	}
}
