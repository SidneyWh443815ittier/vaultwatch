package sender_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestAzureMonitorSender_PostsCorrectPayload(t *testing.T) {
	var received []map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Log-Type") != "VaultWatch" {
			t.Errorf("expected Log-Type VaultWatch, got %s", r.Header.Get("Log-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewAzureMonitorSenderWithURL("ws-id", "shared-key", "VaultWatch", ts.URL)
	if err := s.Send("critical", "lease expiring soon", "lease/abc/123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(received) != 1 {
		t.Fatalf("expected 1 payload entry, got %d", len(received))
	}
	entry := received[0]
	if entry["level"] != "critical" {
		t.Errorf("expected level critical, got %v", entry["level"])
	}
	if entry["lease_id"] != "lease/abc/123" {
		t.Errorf("expected lease_id lease/abc/123, got %v", entry["lease_id"])
	}
	if entry["message"] != "lease expiring soon" {
		t.Errorf("expected message 'lease expiring soon', got %v", entry["message"])
	}
}

func TestAzureMonitorSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewAzureMonitorSenderWithURL("ws-id", "shared-key", "VaultWatch", ts.URL)
	err := s.Send("warning", "test", "lease/xyz")
	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}

func TestAzureMonitorSender_TimestampPresent(t *testing.T) {
	var received []map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewAzureMonitorSenderWithURL("ws-id", "shared-key", "VaultWatch", ts.URL)
	_ = s.Send("info", "ok", "lease/ts")

	if len(received) == 0 {
		t.Fatal("no payload received")
	}
	if ts, ok := received[0]["timestamp"]; !ok || ts == "" {
		t.Error("expected non-empty timestamp field")
	}
}
