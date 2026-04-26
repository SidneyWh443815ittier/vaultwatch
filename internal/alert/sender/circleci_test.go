package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
	"github.com/yourusername/vaultwatch/internal/monitor"
)

func newCircleCISenderWithURL(url, token, projectSlug string) *sender.CircleCISender {
	return sender.NewCircleCISenderWithURL(url, token, projectSlug)
}

func TestCircleCISender_PostsCorrectPayload(t *testing.T) {
	var received map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Circle-Token") != "test-token" {
			t.Errorf("expected Circle-Token header 'test-token', got %q", r.Header.Get("Circle-Token"))
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	s := sender.NewCircleCISenderWithURL(ts.URL, "test-token", "gh/org/repo")
	lease := monitor.LeaseInfo{
		LeaseID: "lease-abc",
		TTL:     30 * time.Minute,
	}

	if err := s.Send("critical", lease); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params, ok := received["parameters"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'parameters' key in payload")
	}
	if params["vault_alert_level"] != "critical" {
		t.Errorf("expected vault_alert_level=critical, got %v", params["vault_alert_level"])
	}
	if params["vault_lease_id"] != "lease-abc" {
		t.Errorf("expected vault_lease_id=lease-abc, got %v", params["vault_lease_id"])
	}
}

func TestCircleCISender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	s := sender.NewCircleCISenderWithURL(ts.URL, "bad-token", "gh/org/repo")
	lease := monitor.LeaseInfo{LeaseID: "lease-xyz", TTL: 5 * time.Minute}

	if err := s.Send("warning", lease); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
