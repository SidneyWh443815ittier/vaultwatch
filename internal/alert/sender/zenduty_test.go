package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestZendutySender_PostsCorrectPayload(t *testing.T) {
	var got map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewZendutySenderWithURL("test-key", ts.URL+"/")
	if err := s.Send("critical", "lease abc expires soon"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["alert_type"] != "critical" {
		t.Errorf("expected alert_type=critical, got %q", got["alert_type"])
	}
	if got["message"] != "lease abc expires soon" {
		t.Errorf("unexpected message: %q", got["message"])
	}
}

func TestZendutySender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewZendutySenderWithURL("test-key", ts.URL+"/")
	if err := s.Send("warning", "msg"); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestZendutySender_AlertTypeMapping(t *testing.T) {
	cases := []struct {
		level    string
		wantType string
	}{
		{"critical", "critical"},
		{"warning", "warning"},
		{"info", "info"},
		{"unknown", "info"},
	}
	for _, tc := range cases {
		var got map[string]string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &got)
			w.WriteHeader(http.StatusOK)
		}))
		s := sender.NewZendutySenderWithURL("k", ts.URL+"/")
		_ = s.Send(tc.level, "msg")
		ts.Close()
		if got["alert_type"] != tc.wantType {
			t.Errorf("level %q: expected %q got %q", tc.level, tc.wantType, got["alert_type"])
		}
	}
}
