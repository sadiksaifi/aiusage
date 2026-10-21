package app

import (
	"fmt"
	"image/color"
	"math"
	"slices"
	"strings"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/ansi"
)

func render(data report, colors bool, width int, dark bool) string {
	theme := fang.DefaultColorScheme(lipgloss.LightDark(dark))
	paint := func(text string, foreground color.Color, bold bool) string {
		if !colors {
			return text
		}
		return lipgloss.NewStyle().Foreground(foreground).Bold(bold).Render(text)
	}
	type section struct {
		name    string
		rows    [][]string
		details []string
	}
	var sections []section
	headers := []string{"WINDOW", "USAGE", "RESETS IN", "RESETS AT"}
	compact := width > 0 && width < 72
	meter := width <= 0 || width >= 60
	if compact {
		headers = []string{"WINDOW", "USAGE", "RESETS"}
	}
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, provider := range []struct {
		name    string
		windows windows
	}{{"Claude", data.Claude}, {"Codex", data.Codex}} {
		block := section{name: provider.name}
		for _, item := range []struct {
			name   string
			window window
		}{{"5-hour", provider.windows.FiveHour}, {"7-day", provider.windows.SevenDay}} {
			if item.window == nil {
				continue
			}
			w := item.window
			usage := ""
			used, ok := w["utilization"].(float64)
			if !ok {
				used, ok = w["used_percent"].(float64)
			}
			if ok {
				shade := lipgloss.LightDark(dark)(lipgloss.Color("#15803D"), lipgloss.Color("#86EFAC"))
				if used >= 90 {
					shade = lipgloss.LightDark(dark)(lipgloss.Color("#B91C1C"), lipgloss.Color("#FCA5A5"))
				} else if used >= 70 {
					shade = lipgloss.LightDark(dark)(lipgloss.Color("#A16207"), lipgloss.Color("#FDE68A"))
				}
				usage = paint(fmt.Sprintf("%g%%", used), shade, true)
				if meter {
					filled := int(math.Round(max(0, min(100, used)) * 12 / 100))
					usage = paint(strings.Repeat("━", filled), shade, false) + paint(strings.Repeat("─", 12-filled), theme.Comment, false) + "  " + usage
				}
			}
			remaining, _ := w["resets_in"].(string)
			date, _ := w["resets_at"].(string)
			reset := paint(remaining, theme.Base, true)
			row := []string{item.name, usage, reset, paint(date, theme.Comment, false)}
			if compact {
				var parts []string
				if remaining != "" {
					parts = append(parts, reset)
				}
				if date != "" {
					parts = append(parts, paint(date, theme.Comment, false))
				}
				row = []string{item.name, usage, strings.Join(parts, "\n")}
			}
			for column, value := range row {
				for _, line := range strings.Split(value, "\n") {
					widths[column] = max(widths[column], ansi.StringWidth(line))
				}
			}
			block.rows = append(block.rows, row)
			var extra []string
			for key, value := range w {
				if value == nil || key == "utilization" || key == "used_percent" || key == "resets_at" || key == "resets_in" {
					continue
				}
				extra = append(extra, key)
			}
			slices.Sort(extra)
			for _, key := range extra {
				block.details = append(block.details, item.name+" "+strings.ReplaceAll(key, "_", " ")+": "+fmt.Sprint(w[key]))
			}
		}
		if len(block.rows) > 0 {
			sections = append(sections, block)
		}
	}
	if len(sections) == 0 {
		return "\n  No usage data available.\n"
	}
	var output []string
	for _, section := range sections {
		t := table.New().Headers(headers...).Rows(section.rows...).
			Border(lipgloss.HiddenBorder()).BorderTop(false).BorderBottom(false).
			BorderLeft(false).BorderRight(false).BorderColumn(false).BorderHeader(false).
			StyleFunc(func(row, column int) lipgloss.Style {
				padding := 3
				if column == len(headers)-1 {
					padding = 0
				}
				style := lipgloss.NewStyle().Width(widths[column] + padding).PaddingRight(padding)
				if row == table.HeaderRow {
					style = style.PaddingBottom(1)
					if colors {
						style = style.Foreground(theme.Comment)
					}
				}
				return style
			})
		body := t.String()
		for _, detail := range section.details {
			body += "\n" + paint(detail, theme.Comment, false)
		}
		output = append(output, "  "+paint(strings.ToUpper(section.name), theme.Title, true)+"\n\n"+lipgloss.NewStyle().MarginLeft(4).Render(body))
	}
	return "\n" + strings.Join(output, "\n\n") + "\n"
}
