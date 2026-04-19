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

func TestMattermostSender_PostsFormattedMessage(t *testing.T) {
	var got map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewMattermostSender(ts.URL)
	a := alert.Alert{
		LeaseID: "secret/data/myapp",
		Level:   "warning",
		TTL:     2 * time.Hour,
		Message: "expiring soon",
	}

	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got["username"] != "vaultwatch" {
		t.Errorf("expected username vaultwatch, got %q", got["username"])
	}
	if got["text"] == "" {
		t.Error("expected non-empty text")
	}
}

func TestMattermostSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewMattermostSender(ts.URL)
	a := alert.Alert{
		LeaseID: "secret/data/myapp",
		Level:   "critical",
		TTL:     10 * time.Minute,
		Message: "lease critical",
	}

	if err := s.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
