package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

// kafkaSender sends alerts to a Kafka REST Proxy endpoint.
type kafkaSender struct {
	url     string
	topic   string
	client  *http.Client
}

type kafkaRecord struct {
	Value kafkaPayload `json:"value"`
}

type kafkaPayload struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	LeaseID string `json:"lease_id"`
	TTL     int64  `json:"ttl_seconds"`
	SentAt  string `json:"sent_at"`
}

type kafkaProduceRequest struct {
	Records []kafkaRecord `json:"records"`
}

// NewKafkaSender creates a Sender that publishes alerts to a Kafka topic via
// the Confluent REST Proxy. url should be the base REST Proxy URL
// (e.g. "http://localhost:8082") and topic is the target topic name.
func NewKafkaSender(url, topic string) Sender {
	return newKafkaSenderWithURL(url, topic, &http.Client{Timeout: 10 * time.Second})
}

func newKafkaSenderWithURL(url, topic string, client *http.Client) Sender {
	return &kafkaSender{url: url, topic: topic, client: client}
}

func (k *kafkaSender) Send(a alert.Alert) error {
	payload := kafkaProduceRequest{
		Records: []kafkaRecord{
			{
				Value: kafkaPayload{
					Level:   string(a.Level),
					Message: a.Message,
					LeaseID: a.LeaseID,
					TTL:     int64(a.TTL.Seconds()),
					SentAt:  time.Now().UTC().Format(time.RFC3339),
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("kafka sender: marshal payload: %w", err)
	}

	endpoint := fmt.Sprintf("%s/topics/%s", k.url, k.topic)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("kafka sender: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/vnd.kafka.json.v2+json")

	resp, err := k.client.Do(req)
	if err != nil {
		return fmt.Errorf("kafka sender: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("kafka sender: unexpected status %d", resp.StatusCode)
	}
	return nil
}
