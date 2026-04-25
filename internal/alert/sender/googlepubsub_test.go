package sender_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/vaultwatch/internal/alert/sender"
)

func TestGooglePubSubSender_PostsCorrectPayload(t *testing.T) {
	var received map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"messageIds":["1"]}`)) 
	}))
	defer srv.Close()

	s := sender.NewGooglePubSubSenderWithBase("my-project", "my-topic", "", srv.URL)
	err := s.Send(context.Background(), "critical", "lease-abc", "lease expiring soon")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msgs, ok := received["messages"].([]interface{})
	if !ok || len(msgs) == 0 {
		t.Fatal("expected at least one message")
	}

	entry := msgs[0].(map[string]interface{})
	dataRaw, _ := base64.StdEncoding.DecodeString(entry["data"].(string))

	var payload map[string]interface{}
	if err := json.Unmarshal(dataRaw, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	if payload["level"] != "critical" {
		t.Errorf("level: got %q, want %q", payload["level"], "critical")
	}
	if payload["lease_id"] != "lease-abc" {
		t.Errorf("lease_id: got %q, want %q", payload["lease_id"], "lease-abc")
	}
	if payload["message"] != "lease expiring soon" {
		t.Errorf("message: got %q, want %q", payload["message"], "lease expiring soon")
	}
	if payload["timestamp"] == "" {
		t.Error("expected non-empty timestamp")
	}

	attrs := entry["attributes"].(map[string]interface{})
	if attrs["level"] != "critical" {
		t.Errorf("attribute level: got %q, want %q", attrs["level"], "critical")
	}
}

func TestGooglePubSubSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	s := sender.NewGooglePubSubSenderWithBase("proj", "topic", "key", srv.URL)
	err := s.Send(context.Background(), "warning", "lease-xyz", "expiring")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestGooglePubSubSender_APIKeyAppendedToURL(t *testing.T) {
	var capturedURL string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"messageIds":["2"]}`))
	}))
	defer srv.Close()

	s := sender.NewGooglePubSubSenderWithBase("proj", "topic", "my-api-key", srv.URL)
	_ = s.Send(context.Background(), "ok", "l1", "msg")

	if capturedURL != "key=my-api-key" {
		t.Errorf("expected key=my-api-key in query, got %q", capturedURL)
	}
}
