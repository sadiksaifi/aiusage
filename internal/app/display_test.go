package app

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return string(data)
}

func TestDisplayOmitsNullsButKeepsZeroAndCountdown(t *testing.T) {
	report := report{Claude: windows{FiveHour: window{"utilization": float64(0), "resets_at": "Thu 22 Oct, 01:02", "resets_in": "1h 2m", "used_dollars": float64(0), "locked_reason": nil}}, Codex: windows{SevenDay: window{"used_percent": float64(100), "resets_at": "Thu 22 Oct, 01:02", "resets_in": "6d 23h 59m"}}}
	for _, width := range []int{30, 35, 40, 80, 120} {
		text := render(report, false, width, true)
		for _, value := range []string{"CLAUDE", "CODEX", "5-hour", "7-day", "0%", "100%", "Thu 22 Oct, 01:02", "1h 2m", "6d 23h 59m", "used dollars: 0"} {
			require.Contains(t, text, value)
		}
		require.NotContains(t, text, "locked reason")
		require.NotContains(t, text, "null")
		require.NotContains(t, text, "\x1b")
		for _, line := range strings.Split(text, "\n") {
			require.LessOrEqual(t, ansi.StringWidth(line), width)
		}
	}
	require.Contains(t, render(report, true, 120, true), "\x1b[")
	require.Contains(t, mustJSON(t, report), `"locked_reason":null`)
}
