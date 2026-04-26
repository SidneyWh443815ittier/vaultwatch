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

func TestKafkaSender_PostsCorrectPayload(t *testing.T) {
	var gotBody []byte
	var gotContentType string
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := sender.NewKafkaSenderWithURL(srv.URL, "vault-alerts", srv.Client())
	a := alert.Alert{
		Level:   alert.LevelCritical,
		Message: "lease expiring soon",
		LeaseID: "lease/abc/123",
		TTL:     5 * time.Minute,
	}

	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/topics/vault-alerts" {
		t.Errorf("expected path /topics/vault-alerts, got %s", gotPath)
	}
	if gotContentType != "application/vnd.kafka.json.v2+json" {
		t.Errorf("unexpected content-type: %s", gotContentType)
	}

	var req map[string]interface{}
	if err := json.Unmarshal(gotBody, &req); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	records, ok := req["records"].([]interface{})
	if !ok || len(records) != 1 {
		t.Fatalf("expected 1 record, got %v", req["records"])
	}
	val := records[0].(map[string]interface{})["value"].(map[string]interface{})
	if val["level"] != string(alert.LevelCritical) {
		t.Errorf("expected level %s, got %v", alert.LevelCritical, val["level"])
	}
	if val["lease_id"] != "lease/abc/123" {
		t.Errorf("expected lease_id lease/abc/123, got %v", val["lease_id"])
	}
	if val["ttl_seconds"] != float64(300) {
		t.Errorf("expected ttl_seconds 300, got %v", val["ttl_seconds"])
	}
}

func TestKafkaSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := sender.NewKafkaSenderWithURL(srv.URL, "vault-alerts", srv.Client())
	err := s.Send(alert.Alert{Level: alert.LevelWarning, TTL: time.Minute})
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
