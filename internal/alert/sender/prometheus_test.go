package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestPrometheusSender_PostsCorrectPayload(t *testing.T) {
	var capturedBody []map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/alerts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected content-type: %s", ct)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewPrometheusSender(ts.URL, "testvaultwatch")
	if err := s.Send("critical", "secret expires soon"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedBody) != 1 {
		t.Fatalf("expected 1 alert in payload, got %d", len(capturedBody))
	}

	alert := capturedBody[0]
	labels, ok := alert["labels"].(map[string]interface{})
	if !ok {
		t.Fatal("labels missing from payload")
	}
	if labels["severity"] != "critical" {
		t.Errorf("expected severity=critical, got %v", labels["severity"])
	}
	if labels["job"] != "testvaultwatch" {
		t.Errorf("expected job=testvaultwatch, got %v", labels["job"])
	}

	annotations, ok := alert["annotations"].(map[string]interface{})
	if !ok {
		t.Fatal("annotations missing from payload")
	}
	if annotations["summary"] != "secret expires soon" {
		t.Errorf("unexpected summary: %v", annotations["summary"])
	}
}

func TestPrometheusSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewPrometheusSender(ts.URL, "")
	if err := s.Send("warning", "lease warning"); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestPrometheusSender_DefaultJobLabel(t *testing.T) {
	var capturedBody []map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewPrometheusSender(ts.URL, "")
	if err := s.Send("info", "test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedBody) < 1 {
		t.Fatal("expected payload")
	}
	labels, _ := capturedBody[0]["labels"].(map[string]interface{})
	if labels["job"] != "vaultwatch" {
		t.Errorf("expected default job=vaultwatch, got %v", labels["job"])
	}
}
