package monitor

import (
	"testing"
	"time"
)

func TestClassifyLease(t *testing.T) {
	thresholds := Thresholds{
		Warning:  24 * time.Hour,
		Critical: 4 * time.Hour,
	}

	cases := []struct {
		name     string
		ttl      time.Duration
		expected LeaseStatus
	}{
		{"expired", 0, LeaseExpired},
		{"negative ttl", -1 * time.Second, LeaseExpired},
		{"critical boundary", 4 * time.Hour, LeaseCritical},
		{"critical below", 1 * time.Hour, LeaseCritical},
		{"warning boundary", 24 * time.Hour, LeaseWarning},
		{"warning range", 12 * time.Hour, LeaseWarning},
		{"ok", 48 * time.Hour, LeaseOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyLease(tc.ttl, thresholds)
			if got != tc.expected {
				t.Errorf("classifyLease(%v) = %d, want %d", tc.ttl, got, tc.expected)
			}
		})
	}
}

func TestClassifyLease_CustomThresholds(t *testing.T) {
	thresholds := Thresholds{
		Warning:  2 * time.Hour,
		Critical: 30 * time.Minute,
	}

	if got := classifyLease(3*time.Hour, thresholds); got != LeaseOK {
		t.Errorf("expected LeaseOK for ttl above warning, got %d", got)
	}
	if got := classifyLease(1*time.Hour, thresholds); got != LeaseWarning {
		t.Errorf("expected LeaseWarning for ttl between thresholds, got %d", got)
	}
	if got := classifyLease(15*time.Minute, thresholds); got != LeaseCritical {
		t.Errorf("expected LeaseCritical for ttl below critical, got %d", got)
	}
}

func TestDefaultThresholds(t *testing.T) {
	if DefaultThresholds.Warning != 24*time.Hour {
		t.Errorf("expected default warning=24h, got %v", DefaultThresholds.Warning)
	}
	if DefaultThresholds.Critical != 4*time.Hour {
		t.Errorf("expected default critical=4h, got %v", DefaultThresholds.Critical)
	}
}

func TestNew(t *testing.T) {
	thresholds := Thresholds{
		Warning:  10 * time.Hour,
		Critical: 2 * time.Hour,
	}
	m := New(nil, thresholds)
	if m == nil {
		t.Fatal("expected non-nil monitor")
	}
	if m.thresholds.Warning != thresholds.Warning {
		t.Errorf("unexpected warning threshold: %v", m.thresholds.Warning)
	}
	if m.thresholds.Critical != thresholds.Critical {
		t.Errorf("unexpected critical threshold: %v", m.thresholds.Critical)
	}
}
