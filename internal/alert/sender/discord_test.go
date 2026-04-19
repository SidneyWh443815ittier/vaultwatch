package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestDiscordSender_PostsFormattedMessage(t *testing.T) {
	var got map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	s := sender.NewDiscordSender(srv.URL)
	a := alert.Alert{
		Level:   alert.LevelCritical,
		LeaseID: "lease/abc",
		Path:    "secret/db",
		TTL:     5 * time.Minute,
	}

	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	embeds, ok := got["embeds"].([]interface{})
	if !ok || len(embeds) == 0 {
		t.Fatal("expected embeds in payload")
	}
	embed := embeds[0].(map[string]interface{})
	if embed["title"] == "" {
		t.Error("expected non-empty title")
	}
}

func TestDiscordSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := sender.NewDiscordSender(srv.URL)
	err := s.Send(alert.Alert{Level: alert.LevelWarning, TTL: time.Minute})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestDiscordColor_Mapping(t *testing.T) {
	var got map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	s := sender.NewDiscordSender(srv.URL)
	_ = s.Send(alert.Alert{Level: alert.LevelCritical, TTL: time.Minute})

	embeds := got["embeds"].([]interface{})
	embed := embeds[0].(map[string]interface{})
	color := int(embed["color"].(float64))
	if color != 0xFF0000 {
		t.Errorf("expected red (0xFF0000) for critical, got %d", color)
	}
}
