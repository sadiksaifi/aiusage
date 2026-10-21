package app

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

func render(data report, colors bool) string {
	paint := func(text, color string, bold bool) string {
		if !colors {
			return text
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(bold).Render(text)
	}
	var sections []string
	for _, provider := range []struct {
		name    string
		windows windows
	}{{"Claude", data.Claude}, {"Codex", data.Codex}} {
		var lines []string
		for _, item := range []struct {
			name   string
			window window
		}{{"5-hour", provider.windows.FiveHour}, {"7-day", provider.windows.SevenDay}} {
			if item.window == nil {
				continue
			}
			w := item.window
			var parts []string
			used, ok := w["utilization"].(float64)
			if !ok {
				used, ok = w["used_percent"].(float64)
			}
			if ok {
				color := "2"
				if used >= 90 {
					color = "1"
				} else if used >= 70 {
					color = "3"
				}
				parts = append(parts, paint(fmt.Sprintf("%g%% used", used), color, false))
			}
			if reset, ok := w["resets_at"].(string); ok {
				parts = append(parts, "resets "+reset)
			}
			if remaining, ok := w["resets_in"].(string); ok {
				if remaining == "now" {
					parts = append(parts, "resets now")
				} else {
					parts = append(parts, "resets in "+remaining)
				}
			}
			var extra []string
			for key, value := range w {
				if value == nil || key == "utilization" || key == "used_percent" || key == "resets_at" || key == "resets_in" {
					continue
				}
				extra = append(extra, key)
			}
			slices.Sort(extra)
			for _, key := range extra {
				parts = append(parts, strings.ReplaceAll(key, "_", " ")+": "+fmt.Sprint(w[key]))
			}
			lines = append(lines, "  "+item.name+": "+strings.Join(parts, " | "))
		}
		if len(lines) > 0 {
			sections = append(sections, paint(provider.name, "6", true)+"\n"+strings.Join(lines, "\n"))
		}
	}
	if len(sections) == 0 {
		return "No usage data available."
	}
	return strings.Join(sections, "\n\n")
}
