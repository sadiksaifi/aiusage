package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type codexWindow struct {
	UsedPercent *float64 `json:"used_percent"`
	Seconds     int64    `json:"limit_window_seconds"`
	ResetAt     *int64   `json:"reset_at"`
}

type client struct {
	http      *http.Client
	claudeURL string
	codexURL  string
}

func newClient() client {
	return client{
		http:      &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
		claudeURL: "https://api.anthropic.com/api/oauth/usage",
		codexURL:  "https://chatgpt.com/backend-api/wham/usage",
	}
}

func (c client) get(ctx context.Context, url string, auth login, headers map[string]string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+auth.token)
	req.Header.Set("Accept", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	response, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("usage request failed: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		hint := ""
		switch response.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			hint = "; sign in again using the provider CLI (token read from " + auth.source + ")"
		case http.StatusTooManyRequests:
			hint = "; rate limited; retry " + retryAfter(response.Header.Get("Retry-After"))
		}
		return fmt.Errorf("usage API returned HTTP %d%s", response.StatusCode, hint)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(target); err != nil {
		return fmt.Errorf("invalid usage response: %w", err)
	}
	return nil
}

// retryAfter reads a Retry-After value, which is either delay seconds or an HTTP date.
func retryAfter(value string) string {
	var wait time.Duration
	if seconds, err := strconv.Atoi(value); err == nil {
		wait = time.Duration(seconds) * time.Second
	} else if at, err := http.ParseTime(value); err == nil {
		wait = time.Until(at)
	}
	if wait <= 0 {
		return "later"
	}
	return "in " + countdown(wait)
}

// collect fetches each provider independently, so one failure leaves the other's usage intact.
func (c client) collect(ctx context.Context, auth credentials, now time.Time) report {
	var result report
	var wg sync.WaitGroup
	wg.Go(func() { result.Claude = c.claude(ctx, auth.claude, now) })
	wg.Go(func() { result.Codex = c.codex(ctx, auth.codex, now) })
	wg.Wait()
	return result
}

func (c client) claude(ctx context.Context, auth login, now time.Time) windows {
	if auth.err != nil {
		return failed(auth.err)
	}
	var data struct {
		FiveHour window `json:"five_hour"`
		SevenDay window `json:"seven_day"`
	}
	if err := c.get(ctx, c.claudeURL, auth, map[string]string{"anthropic-beta": "oauth-2025-04-20"}, &data); err != nil {
		return failed(err)
	}
	for _, w := range []window{data.FiveHour, data.SevenDay} {
		if w == nil {
			continue
		}
		raw, ok := w["resets_at"].(string)
		if !ok {
			w["resets_in"] = nil
			continue
		}
		reset, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return failed(fmt.Errorf("invalid reset timestamp"))
		}
		for key, value := range resetFields(reset, now) {
			w[key] = value
		}
	}
	return windows{FiveHour: data.FiveHour, SevenDay: data.SevenDay}
}

func (c client) codex(ctx context.Context, auth login, now time.Time) windows {
	if auth.err != nil {
		return failed(auth.err)
	}
	var data struct {
		RateLimit struct {
			Primary   *codexWindow `json:"primary_window"`
			Secondary *codexWindow `json:"secondary_window"`
		} `json:"rate_limit"`
	}
	if err := c.get(ctx, c.codexURL, auth, map[string]string{"ChatGPT-Account-Id": auth.account}, &data); err != nil {
		return failed(err)
	}
	var result windows
	for _, raw := range []*codexWindow{data.RateLimit.Primary, data.RateLimit.Secondary} {
		if raw == nil {
			continue
		}
		w := window{"used_percent": nil, "resets_at": nil, "resets_in": nil}
		if raw.UsedPercent != nil {
			w["used_percent"] = *raw.UsedPercent
		}
		if raw.ResetAt != nil {
			for key, value := range resetFields(time.Unix(*raw.ResetAt, 0), now) {
				w[key] = value
			}
		}
		switch raw.Seconds {
		case 18000:
			result.FiveHour = w
		case 604800:
			result.SevenDay = w
		}
	}
	return result
}
