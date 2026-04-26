package sender_test

import (
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestSquadcastSender_ImplementsInterface(t *testing.T) {
	var _ alert.Sender = sender.NewSquadcastSender("https://example.com/webhook")
}
