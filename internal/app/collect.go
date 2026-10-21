package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func (c client) get(ctx context.Context, url, token string, headers map[string]string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
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
		if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
			hint = "; sign in again using the provider CLI"
		}
		return fmt.Errorf("usage API returned HTTP %d%s", response.StatusCode, hint)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(target); err != nil {
		return fmt.Errorf("invalid usage response: %w", err)
	}
	return nil
}

func (c client) collect(ctx context.Context, auth credentials, now time.Time) (report, error) {
	var result report
	var claudeErr, codexErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		claudeErr = c.get(ctx, c.claudeURL, auth.claudeToken, map[string]string{"anthropic-beta": "oauth-2025-04-20"}, &result.Claude)
		if claudeErr != nil {
			return
		}
		for _, w := range []window{result.Claude.FiveHour, result.Claude.SevenDay} {
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
				claudeErr = fmt.Errorf("invalid reset timestamp")
				return
			}
			for key, value := range resetFields(reset, now) {
				w[key] = value
			}
		}
	}()
	go func() {
		defer wg.Done()
		var data struct {
			RateLimit struct {
				Primary   *codexWindow `json:"primary_window"`
				Secondary *codexWindow `json:"secondary_window"`
			} `json:"rate_limit"`
		}
		codexErr = c.get(ctx, c.codexURL, auth.codexToken, map[string]string{"ChatGPT-Account-Id": auth.accountID}, &data)
		if codexErr != nil {
			return
		}
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
				result.Codex.FiveHour = w
			case 604800:
				result.Codex.SevenDay = w
			}
		}
	}()
	wg.Wait()
	if claudeErr != nil {
		return report{}, fmt.Errorf("claude: %w", claudeErr)
	}
	if codexErr != nil {
		return report{}, fmt.Errorf("codex: %w", codexErr)
	}
	return result, nil
}
