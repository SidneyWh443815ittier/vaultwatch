package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// awsSNSPayload represents the JSON body sent to an AWS SNS HTTP subscription endpoint
// or a compatible SNS-over-HTTP proxy.
type awsSNSPayload struct {
	Type      string `json:"Type"`
	Subject   string `json:"Subject"`
	Message   string `json:"Message"`
	Timestamp string `json:"Timestamp"`
}

// AWSSNSSender sends alert notifications to an AWS SNS-compatible HTTP endpoint.
// It is useful for SNS HTTP/HTTPS subscriptions or local testing proxies.
type AWSSNSSender struct {
	endpointURL string
	client      *http.Client
}

// NewAWSSNSSender creates an AWSSNSSender that posts to the given SNS endpoint URL.
// The endpoint should be an SNS subscription confirmation or delivery URL.
func NewAWSSNSSender(endpointURL string) *AWSSNSSender {
	return newAWSSNSSenderWithURL(endpointURL, &http.Client{Timeout: 10 * time.Second})
}

func newAWSSNSSenderWithURL(endpointURL string, client *http.Client) *AWSSNSSender {
	return &AWSSNSSender{
		endpointURL: endpointURL,
		client:      client,
	}
}

// Send delivers an alert for the given lease classification to the SNS endpoint.
// The notification level is embedded in the Subject field for easy filtering.
func (s *AWSSNSSender) Send(lease monitor.LeaseInfo, level AlertLevel) error {
	subject := fmt.Sprintf("[VaultWatch][%s] Lease expiring: %s", level, lease.LeaseID)
	message := fmt.Sprintf(
		"Lease ID : %s\nPath     : %s\nExpires  : %s\nTTL      : %s\nLevel    : %s",
		lease.LeaseID,
		lease.Path,
		lease.ExpireTime.UTC().Format(time.RFC3339),
		time.Until(lease.ExpireTime).Round(time.Second),
		level,
	)

	payload := awsSNSPayload{
		Type:      "Notification",
		Subject:   subject,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("awssns: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.endpointURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("awssns: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("awssns: unexpected status %d", resp.StatusCode)
	}
	return nil
}
