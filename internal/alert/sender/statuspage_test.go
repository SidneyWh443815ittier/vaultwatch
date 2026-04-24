package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestStatuspageSender_PostsCorrectPayload(t *testing.T) {
	var gotBody map[string]map[string]string
	var gotPath string
	var gotAuth string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewStatuspageSenderWithURL("test-key", "page123", "comp456", ts.URL)
	err := s.Send(alert.Message{
		Lease:   "secret/my-app",
		Level:   alert.LevelCritical,
		Message: "Lease expiring soon",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantPath := "/pages/page123/components/comp456"
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}
	if gotAuth != "OAuth test-key" {
		t.Errorf("auth = %q, want %q", gotAuth, "OAuth test-key")
	}
	if gotBody["component"]["status"] != "major_outage" {
		t.Errorf("status = %q, want %q", gotBody["component"]["status"], "major_outage")
	}
}

func TestStatuspageSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	s := sender.NewStatuspageSenderWithURL("bad-key", "page123", "comp456", ts.URL)
	err := s.Send(alert.Message{Level: alert.LevelWarning})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStatuspageSender_StatusMapping(t *testing.T) {
	cases := []struct {
		level  alert.Level
		want   string
	}{
		{alert.LevelCritical, "major_outage"},
		{alert.LevelWarning, "degraded_performance"},
		{alert.LevelOK, "operational"},
	}

	for _, tc := range cases {
		var gotBody map[string]map[string]string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(data, &gotBody)
			w.WriteHeader(http.StatusOK)
		}))

		s := sender.NewStatuspageSenderWithURL("key", "p", "c", ts.URL)
		_ = s.Send(alert.Message{Level: tc.level})
		ts.Close()

		if got := gotBody["component"]["status"]; got != tc.want {
			t.Errorf("level %v: status = %q, want %q", tc.level, got, tc.want)
		}
	}
}
