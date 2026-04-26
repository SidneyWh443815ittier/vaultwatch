package sender_test

import (
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

// TestMatrixSender_Interface ensures matrixSender satisfies the Sender interface.
func TestMatrixSender_Interface(t *testing.T) {
	var _ sender.Sender = sender.NewMatrixSender("!room:example.com", "token")
}
