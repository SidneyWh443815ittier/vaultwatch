package alert

import (
	"fmt"
	"log"
	"time"

	"github.com/user/vaultwatch/internal/monitor"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelCritical Level = "CRITICAL"
	LevelWarning  Level = "WARNING"
	LevelInfo     Level = "INFO"
)

// Alert holds information about a lease expiration alert.
type Alert struct {
	LeaseID   string
	Level     Level
	ExpiresIn time.Duration
	ExpireAt  time.Time
	Message   string
}

// Sender is the interface for alert delivery backends.
type Sender interface {
	Send(alert Alert) error
}

// Notifier evaluates classified leases and dispatches alerts.
type Notifier struct {
	senders []Sender
}

// New creates a Notifier with the provided Sender backends.
func New(senders ...Sender) *Notifier {
	return &Notifier{senders: senders}
}

// Notify evaluates a LeaseStatus and sends an alert if warranted.
// Returns the number of senders that succeeded.
func (n *Notifier) Notify(status monitor.LeaseStatus) int {
	if status.Classification == monitor.ClassOK {
		return 0
	}

	level := classificationToLevel(status.Classification)
	a := Alert{
		LeaseID:   status.LeaseID,
		Level:     level,
		ExpiresIn: status.TTL,
		ExpireAt:  time.Now().Add(status.TTL),
		Message: fmt.Sprintf("[%s] Lease %s expires in %s",
			level, status.LeaseID, status.TTL.Round(time.Second)),
	}

	sent := 0
	for _, s := range n.senders {
		if err := s.Send(a); err != nil {
			log.Printf("alert sender error: %v", err)
			continue
		}
		sent++
	}
	return sent
}

func classificationToLevel(c monitor.Classification) Level {
	switch c {
	case monitor.ClassCritical:
		return LevelCritical
	case monitor.ClassWarning:
		return LevelWarning
	default:
		return LevelInfo
	}
}
