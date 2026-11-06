package app

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type window map[string]any

// windows is one provider's usage, or the error that replaced it.
type windows struct {
	FiveHour  window  `json:"five_hour"`
	SevenDay  window  `json:"seven_day"`
	Error     *string `json:"error"`
	signedOut bool
}

func failed(err error) windows {
	message := err.Error()
	return windows{Error: &message, signedOut: errors.Is(err, errSignedOut)}
}

type report struct {
	Claude windows `json:"claude"`
	Codex  windows `json:"codex"`
}

// err reports why nothing can be shown when every provider failed.
func (r report) err() error {
	if r.Claude.Error == nil || r.Codex.Error == nil {
		return nil
	}
	return fmt.Errorf("no usage to show\nClaude: %s\nCodex: %s", *r.Claude.Error, *r.Codex.Error)
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
