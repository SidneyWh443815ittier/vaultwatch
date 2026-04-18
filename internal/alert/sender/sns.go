package sender

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/your-org/vaultwatch/internal/alert"
)

type snsSender struct {
	topicARN string
	region   string
	apiURL   string
	client   *http.Client
}

type snsPublishRequest struct {
	TopicARN string `json:"TopicArn"`
	Message  string `json:"Message"`
	Subject  string `json:"Subject"`
}

// NewSNSSender creates a sender that publishes alerts to an AWS SNS topic via HTTP.
func NewSNSSender(topicARN, region string) *snsSender {
	url := fmt.Sprintf("https://sns.%s.amazonaws.com/", region)
	return newSNSSenderWithURL(topicARN, region, url)
}

func newSNSSenderWithURL(topicARN, region, apiURL string) *snsSender {
	return &snsSender{
		topicARN: topicARN,
		region:   region,
		apiURL:   apiURL,
		client:   &http.Client{},
	}
}

func (s *snsSender) Send(a alert.Alert) error {
	payload := snsPublishRequest{
		TopicARN: s.topicARN,
		Message:  fmt.Sprintf("[%s] Lease %s expires in %s", strings.ToUpper(a.Level), a.LeaseID, a.TTL),
		Subject:  fmt.Sprintf("VaultWatch %s Alert", strings.Title(a.Level)),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("sns: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.apiURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("sns: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sns: unexpected status %d", resp.StatusCode)
	}
	return nil
}
