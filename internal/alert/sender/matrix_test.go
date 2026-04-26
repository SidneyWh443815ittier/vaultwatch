package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestMatrixSender_PostsFormattedMessage(t *testing.T) {
	var gotBody map[string]string
	var gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := sender.NewMatrixSenderWithBase(srv.URL, "!roomid:example.com", "tok123")
	if err := s.Send("critical", "token expires soon"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAuth != "Bearer tok123" {
		t.Errorf("expected Authorization header 'Bearer tok123', got %q", gotAuth)
	}
	if gotBody["msgtype"] != "m.text" {
		t.Errorf("expected msgtype 'm.text', got %q", gotBody["msgtype"])
	}
	if !strings.Contains(gotBody["body"], "critical") || !strings.Contains(gotBody["body"], "token expires soon") {
		t.Errorf("body missing expected content: %q", gotBody["body"])
	}
}

func TestMatrixSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	s := sender.NewMatrixSenderWithBase(srv.URL, "!roomid:example.com", "badtoken")
	err := s.Send("warning", "lease warning")
	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected error to mention status 403, got: %v", err)
	}
}

func TestMatrixSender_URLContainsRoomID(t *testing.T) {
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	roomID := "!testroomid:matrix.org"
	s := sender.NewMatrixSenderWithBase(srv.URL, roomID, "token")
	_ = s.Send("info", "test")

	if !strings.Contains(gotPath, roomID) {
		t.Errorf("expected URL path to contain room ID %q, got %q", roomID, gotPath)
	}
}
