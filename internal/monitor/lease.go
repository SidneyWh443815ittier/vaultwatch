package monitor

import (
	"log"
	"time"

	"github.com/vaultwatch/internal/vault"
)

// LeaseStatus represents the expiration state of a lease.
type LeaseStatus int

const (
	LeaseOK LeaseStatus = iota
	LeaseWarning
	LeaseCritical
	LeaseExpired
)

// LeaseInfo holds metadata about a monitored lease.
type LeaseInfo struct {
	LeaseID   string
	TTL       time.Duration
	Status    LeaseStatus
	CheckedAt time.Time
}

// Thresholds defines warning and critical TTL thresholds.
type Thresholds struct {
	Warning  time.Duration
	Critical time.Duration
}

// DefaultThresholds provides sensible default alert thresholds.
var DefaultThresholds = Thresholds{
	Warning:  24 * time.Hour,
	Critical: 4 * time.Hour,
}

// Monitor checks lease expiration against configured thresholds.
type Monitor struct {
	client     *vault.Client
	thresholds Thresholds
}

// New creates a new Monitor with the given Vault client and thresholds.
func New(client *vault.Client, thresholds Thresholds) *Monitor {
	return &Monitor{
		client:     client,
		thresholds: thresholds,
	}
}

// CheckLease looks up a lease by ID and evaluates its expiration status.
func (m *Monitor) CheckLease(leaseID string) (*LeaseInfo, error) {
	ttl, err := m.client.LookupLease(leaseID)
	if err != nil {
		return nil, err
	}

	info := &LeaseInfo{
		LeaseID:   leaseID,
		TTL:       ttl,
		CheckedAt: time.Now(),
		Status:    classifyLease(ttl, m.thresholds),
	}

	log.Printf("[monitor] lease %s TTL=%s status=%d", leaseID, ttl, info.Status)
	return info, nil
}

// classifyLease returns the LeaseStatus based on remaining TTL and thresholds.
func classifyLease(ttl time.Duration, t Thresholds) LeaseStatus {
	switch {
	case ttl <= 0:
		return LeaseExpired
	case ttl <= t.Critical:
		return LeaseCritical
	case ttl <= t.Warning:
		return LeaseWarning
	default:
		return LeaseOK
	}
}
