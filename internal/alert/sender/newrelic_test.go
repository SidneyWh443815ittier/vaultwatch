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

func TestNewRelicSender_PostsCorrectPayload(t *testing.T) {
	var got map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("missing api key header")
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewNewRelicSenderWithURL("test-key", ts.URL)
	err := s.Send(alert.Notification{
		Level:   alert.LevelCritical,
		Message: "lease expiring soon",
		LeaseID: "lease/abc",
		TTL:     30 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	event, ok := got["event"].(map[string]interface{})
	if !ok {
		t.Fatal("missing event field")
	}
	if event["eventType"] != "VaultLeaseAlert" {
		t.Errorf("unexpected eventType: %v", event["eventType"])
	}
	if event["level"] != string(alert.LevelCritical) {
		t.Errorf("unexpected level: %v", event["level"])
	}
}

func TestNewRelicSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	s := sender.NewNewRelicSenderWithURL("bad-key", ts.URL)
	err := s.Send(alert.Notification{
		Level:   alert.LevelWarning,
		Message: "warn",
		LeaseID: "lease/xyz",
		TTL:     5 * time.Minute,
	})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
