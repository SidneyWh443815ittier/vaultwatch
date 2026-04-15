package sender_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestLogSender_WritesFormattedLine(t *testing.T) {
	var buf bytes.Buffer
	s := &sender.LogSender{Out: &buf}

	if err := s.Send("CRITICAL", "lease abc123 expires in 5m"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := buf.String()
	if !strings.Contains(line, "[CRITICAL]") {
		t.Errorf("expected [CRITICAL] in output, got: %s", line)
	}
	if !strings.Contains(line, "lease abc123 expires in 5m") {
		t.Errorf("expected message in output, got: %s", line)
	}
}

func TestWebhookSender_PostsJSON(t *testing.T) {
	var received map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := sender.NewWebhookSender(srv.URL)
	if err := s.Send("WARNING", "lease xyz expires in 1h"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["level"] != "WARNING" {
		t.Errorf("expected level WARNING, got %s", received["level"])
	}
	if received["message"] != "lease xyz expires in 1h" {
		t.Errorf("unexpected message: %s", received["message"])
	}
}

func TestWebhookSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := sender.NewWebhookSender(srv.URL)
	if err := s.Send("CRITICAL", "something"); err == nil {
		t.Error("expected error for non-2xx status, got nil")
	}
}
