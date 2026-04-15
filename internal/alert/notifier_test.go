package alert_test

import (
	"errors"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/alert"
	"github.com/user/vaultwatch/internal/monitor"
)

// mockSender records sent alerts and can simulate errors.
type mockSender struct {
	Sent  []alert.Alert
	fail  bool
}

func (m *mockSender) Send(a alert.Alert) error {
	if m.fail {
		return errors.New("send failed")
	}
	m.Sent = append(m.Sent, a)
	return nil
}

func TestNotify_OKClassificationSendsNothing(t *testing.T) {
	s := &mockSender{}
	n := alert.New(s)
	status := monitor.LeaseStatus{
		LeaseID:        "lease/ok",
		TTL:            48 * time.Hour,
		Classification: monitor.ClassOK,
	}
	if got := n.Notify(status); got != 0 {
		t.Errorf("expected 0 sends, got %d", got)
	}
	if len(s.Sent) != 0 {
		t.Errorf("expected no alerts, got %d", len(s.Sent))
	}
}

func TestNotify_CriticalClassificationSendsAlert(t *testing.T) {
	s := &mockSender{}
	n := alert.New(s)
	status := monitor.LeaseStatus{
		LeaseID:        "lease/critical",
		TTL:            30 * time.Minute,
		Classification: monitor.ClassCritical,
	}
	if got := n.Notify(status); got != 1 {
		t.Errorf("expected 1 send, got %d", got)
	}
	if len(s.Sent) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(s.Sent))
	}
	if s.Sent[0].Level != alert.LevelCritical {
		t.Errorf("expected CRITICAL level, got %s", s.Sent[0].Level)
	}
}

func TestNotify_SenderFailureCountedCorrectly(t *testing.T) {
	good := &mockSender{}
	bad := &mockSender{fail: true}
	n := alert.New(bad, good)
	status := monitor.LeaseStatus{
		LeaseID:        "lease/warn",
		TTL:            2 * time.Hour,
		Classification: monitor.ClassWarning,
	}
	if got := n.Notify(status); got != 1 {
		t.Errorf("expected 1 successful send, got %d", got)
	}
}

func TestNotify_WarningLevel(t *testing.T) {
	s := &mockSender{}
	n := alert.New(s)
	status := monitor.LeaseStatus{
		LeaseID:        "lease/warning",
		TTL:            3 * time.Hour,
		Classification: monitor.ClassWarning,
	}
	n.Notify(status)
	if len(s.Sent) == 0 {
		t.Fatal("expected at least one alert")
	}
	if s.Sent[0].Level != alert.LevelWarning {
		t.Errorf("expected WARNING level, got %s", s.Sent[0].Level)
	}
}
