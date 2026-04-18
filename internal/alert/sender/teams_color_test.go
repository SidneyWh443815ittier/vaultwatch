package sender

import (
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert"
)

func TestTeamsColor(t *testing.T) {
	cases := []struct {
		level    alert.Level
		expected string
	}{
		{alert.LevelCritical, "FF0000"},
		{alert.LevelWarning, "FFA500"},
		{alert.LevelOK, "00FF00"},
	}
	for _, tc := range cases {
		if got := teamsColor(tc.level); got != tc.expected {
			t.Errorf("teamsColor(%s) = %s, want %s", tc.level, got, tc.expected)
		}
	}
}
