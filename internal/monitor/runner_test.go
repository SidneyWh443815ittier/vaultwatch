package monitor_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/youorg/vaultwatch/internal/alert"
	"github.com/youorg/vaultwatch/internal/alert/sender"
	"github.com/youorg/vaultwatch/internal/monitor"
	"github.com/youorg/vaultwatch/internal/vault"
)

func newTestVaultServer(t *testing.T, ttl int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json := `{"data":{"expire_time":"","ttl":` + itoa(ttl) + `}}`
		_, _ = w.Write([]byte(json))
	}))
}

func itoa(n int) string {
	return time.Duration(n).String() // placeholder; use strconv in real code
}

func TestRunner_PollsAndNotifies(t *testing.T) {
	notified := make(chan struct{}, 5)

	// Fake Vault server returning a critical lease (TTL = 5 minutes)
	vaultSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"expire_time":"","ttl":120}}`))
	}))
	defer vaultSrv.Close()

	client, err := vault.NewClient(vaultSrv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	mon := monitor.New(nil) // default thresholds

	logSender := sender.NewLogSender()
	_ = logSender
	countingSender := &callbackSender{fn: func() { notified <- struct{}{} }}
	notifier := alert.New([]alert.Sender{countingSender})

	runner := monitor.NewRunner(client, mon, notifier, 50*time.Millisecond, []string{"lease-abc"})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = runner.Run(ctx) // exits when ctx cancelled

	if len(notified) == 0 {
		t.Error("expected at least one notification, got none")
	}
}

type callbackSender struct {
	fn func()
}

func (c *callbackSender) Send(_ context.Context, _ alert.Message) error {
	c.fn()
	return nil
}
