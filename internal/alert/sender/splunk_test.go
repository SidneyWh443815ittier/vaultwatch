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

func TestSplunkSender_PostsCorrectPayload(t *testing.T) {
	var got map[string]interface{}
	var authHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewSplunkSenderWithURL(ts.URL, "mytoken")
	err := s.Send(alert.Alert{
		LeaseID: "lease/abc",
		Level:   alert.Warning,
		TTL:     2 * time.Hour,
		At:      time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if authHeader != "Splunk mytoken" {
		t.Errorf("expected Splunk auth header, got %q", authHeader)
	}
	event, ok := got["event"].(map[string]interface{})
	if !ok {
		t.Fatal("missing event field")
	}
	if event["lease_id"] != "lease/abc" {
		t.Errorf("unexpected lease_id: %v", event["lease_id"])
	}
	if got["sourcetype"] != "vaultwatch" {
		t.Errorf("unexpected sourcetype: %v", got["sourcetype"])
	}
}

func TestSplunkSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	s := sender.NewSplunkSenderWithURL(ts.URL, "badtoken")
	err := s.Send(alert.Alert{
		LeaseID: "lease/xyz",
		Level:   alert.Critical,
		TTL:     30 * time.Minute,
		At:      time.Now(),
	})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
