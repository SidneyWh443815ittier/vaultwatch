package sender_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

// TestMatrixSender_LevelInBody verifies that the alert level is embedded in the message body.
func TestMatrixSender_LevelInBody(t *testing.T) {
	levels := []string{"ok", "warning", "critical"}

	for _, level := range levels {
		level := level
		t.Run(level, func(t *testing.T) {
			var captured string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				buf := make([]byte, 1024)
				n, _ := r.Body.Read(buf)
				captured = string(buf[:n])
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			s := sender.NewMatrixSenderWithBase(srv.URL, "!r:example.com", "tok")
			if err := s.Send(level, "some message"); err != nil {
				t.Fatalf("Send(%q) error: %v", level, err)
			}
			expected := fmt.Sprintf("[%s]", level)
			if len(captured) == 0 {
				t.Fatal("no body captured")
			}
			if !containsStr(captured, expected) {
				t.Errorf("expected body to contain %q, got: %s", expected, captured)
			}
		})
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
