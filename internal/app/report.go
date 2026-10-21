package app

import (
	"fmt"
	"strings"
	"time"
)

type window map[string]any

type windows struct {
	FiveHour window `json:"five_hour"`
	SevenDay window `json:"seven_day"`
}

type report struct {
	Claude windows `json:"claude"`
	Codex  windows `json:"codex"`
}

func countdown(remaining time.Duration) string {
	if remaining <= 0 {
		return "now"
	}
	if remaining < time.Minute {
		return "less than 1m"
	}
	minutes := int64(remaining / time.Minute)
	var parts []string
	for _, unit := range []struct {
		size   int64
		suffix string
	}{{1440, "d"}, {60, "h"}, {1, "m"}} {
		if count := minutes / unit.size; count > 0 {
			parts = append(parts, fmt.Sprintf("%d%s", count, unit.suffix))
		}
		minutes %= unit.size
	}
	return strings.Join(parts, " ")
}

func resetFields(reset, now time.Time) window {
	return window{"resets_at": reset.In(now.Location()).Format("Mon 2 Jan, 15:04"), "resets_in": countdown(reset.Sub(now))}
}
