package sender_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
	"github.com/your-org/vaultwatch/internal/alert/sender"
)

func TestSNSSender_PostsCorrectPayload(t *testing.T) {
	var got map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewSNSSenderWithURL("arn:aws:sns:us-east-1:123456789012:alerts", "us-east-1", ts.URL)
	a := alert.Alert{
		LeaseID: "lease/abc",
		Level:   "critical",
		TTL:     30 * time.Minute,
	}

	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got["TopicArn"] != "arn:aws:sns:us-east-1:123456789012:alerts" {
		t.Errorf("unexpected TopicArn: %q", got["TopicArn"])
	}
	if !strings.Contains(got["Message"], "lease/abc") {
		t.Errorf("message missing lease ID: %q", got["Message"])
	}
	if !strings.Contains(got["Subject"], "Critical") {
		t.Errorf("subject missing level: %q", got["Subject"])
	}
}

func TestSNSSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewSNSSenderWithURL("arn:aws:sns:us-east-1:123:topic", "us-east-1", ts.URL)
	err := s.Send(alert.Alert{Level: "warning", TTL: time.Minute})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
