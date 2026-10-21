package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return string(data)
}

func TestDisplayOmitsNullsButKeepsZeroAndCountdown(t *testing.T) {
	report := report{Claude: windows{FiveHour: window{"utilization": float64(0), "resets_at": "Thu 22 Oct, 01:02", "resets_in": "1h 2m", "used_dollars": float64(0), "locked_reason": nil}}, Codex: windows{SevenDay: window{"used_percent": float64(94)}}}
	text := render(report, false)
	require.Equal(t, "Claude\n  5-hour: 0% used | resets Thu 22 Oct, 01:02 | resets in 1h 2m | used dollars: 0\n\nCodex\n  7-day: 94% used", text)
	require.NotContains(t, text, "null")
	require.NotContains(t, text, "\x1b")
	require.Contains(t, render(report, true), "\x1b[")
	require.Contains(t, mustJSON(t, report), `"locked_reason":null`)
}
