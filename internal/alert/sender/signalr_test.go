package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestSignalRSender_PostsFormattedMessage(t *testing.T) {
	var gotBody map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewSignalRSenderWithURL(ts.URL, ts.Client())
	if err := s.Send("CRITICAL", "lease expires soon"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "[CRITICAL] lease expires soon"
	if gotBody["text"] != want {
		t.Errorf("text = %q, want %q", gotBody["text"], want)
	}
}

func TestSignalRSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewSignalRSenderWithURL(ts.URL, ts.Client())
	if err := s.Send("WARNING", "some message"); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
