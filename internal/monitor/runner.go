package monitor

import (
	"context"
	"log"
	"time"

	"github.com/youorg/vaultwatch/internal/alert"
	"github.com/youorg/vaultwatch/internal/vault"
)

// Runner periodically polls Vault leases and triggers alerts.
type Runner struct {
	client   *vault.Client
	monitor  *LeaseMonitor
	notifier *alert.Notifier
	interval time.Duration
	leaseIDs []string
}

// NewRunner creates a Runner with the given dependencies.
func NewRunner(client *vault.Client, monitor *LeaseMonitor, notifier *alert.Notifier, interval time.Duration, leaseIDs []string) *Runner {
	return &Runner{
		client:   client,
		monitor:  monitor,
		notifier: notifier,
		interval: interval,
		leaseIDs: leaseIDs,
	}
}

// Run starts the polling loop, blocking until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	// Poll immediately on start before waiting for the first tick.
	r.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			r.poll(ctx)
		}
	}
}

func (r *Runner) poll(ctx context.Context) {
	for _, id := range r.leaseIDs {
		lease, err := r.client.LookupLease(ctx, id)
		if err != nil {
			log.Printf("[vaultwatch] failed to look up lease %s: %v", id, err)
			continue
		}

		result := r.monitor.Classify(lease)
		if err := r.notifier.Notify(ctx, result); err != nil {
			log.Printf("[vaultwatch] failed to send alert for lease %s: %v", id, err)
		}
	}
}
