package sender

import (
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert"
)

func TestDiscordColor(t *testing.T) {
	tests := []struct {
		level    alert.Level
		expected int
	}{
		{alert.LevelCritical, 0xFF0000},
		{alert.LevelWarning, 0xFFA500},
		{alert.LevelOK, 0x00FF00},
	}
	for _, tt := range tests {
		got := discordColor(tt.level)
		if got != tt.expected {
			t.Errorf("discordColor(%v) = %d, want %d", tt.level, got, tt.expected)
		}
	}
}
