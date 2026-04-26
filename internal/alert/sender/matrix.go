package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultMatrixBaseURL = "https://matrix-client.matrix.org"

type matrixSender struct {
	baseURL     string
	roomID      string
	accessToken string
	client      *http.Client
}

type matrixMessage struct {
	MsgType string `json:"msgtype"`
	Body    string `json:"body"`
}

// NewMatrixSender creates a sender that posts alerts to a Matrix room.
func NewMatrixSender(roomID, accessToken string) Sender {
	return newMatrixSenderWithBase(defaultMatrixBaseURL, roomID, accessToken)
}

func newMatrixSenderWithBase(baseURL, roomID, accessToken string) Sender {
	return &matrixSender{
		baseURL:     baseURL,
		roomID:      roomID,
		accessToken: accessToken,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *matrixSender) Send(level, message string) error {
	payload := matrixMessage{
		MsgType: "m.text",
		Body:    fmt.Sprintf("[%s] %s", level, message),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("matrix: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/send/m.room.message", s.baseURL, s.roomID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("matrix: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("matrix: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("matrix: unexpected status %d", resp.StatusCode)
	}
	return nil
}
