package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultwatch/internal/alert/sender"
)

func TestSlackSender_PostsFormattedMessage(t *testing.T) {
	var gotBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewSlackSenderWithURL(ts.URL, ts.Client())
	if err := s.Send("CRITICAL", "lease expires soon"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}

	if !strings.Contains(payload["text"], "CRITICAL") {
		t.Errorf("expected level in text, got: %s", payload["text"])
	}
	if !strings.Contains(payload["text"], "lease expires soon") {
		t.Errorf("expected message in text, got: %s", payload["text"])
	}
}

func TestSlackSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewSlackSenderWithURL(ts.URL, ts.Client())
	err := s.Send("WARNING", "some message")
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
