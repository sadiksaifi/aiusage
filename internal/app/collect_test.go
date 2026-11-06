package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCollectPreservesFieldsAndAddsResetCountdown(t *testing.T) {
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/claude":
			require.Equal(t, "Bearer claude-token", r.Header.Get("Authorization"))
			require.Equal(t, "oauth-2025-04-20", r.Header.Get("anthropic-beta"))
			_, _ = w.Write([]byte(`{"five_hour":{"utilization":84,"resets_at":"2026-10-22T01:02:00.123+00:00","used_dollars":0,"locked_reason":null},"seven_day":null}`))
		case "/codex":
			require.Equal(t, "Bearer codex-token", r.Header.Get("Authorization"))
			require.Equal(t, "account", r.Header.Get("ChatGPT-Account-Id"))
			_, _ = w.Write([]byte(`{"rate_limit":{"primary_window":{"used_percent":20,"limit_window_seconds":604800,"reset_at":1792713720},"secondary_window":null}}`))
		}
	}))
	defer server.Close()
	c := client{http: server.Client(), claudeURL: server.URL + "/claude", codexURL: server.URL + "/codex"}
	report := c.collect(context.Background(), stored(login{token: "claude-token"}), stored(login{token: "codex-token", account: "account"}), now)
	require.NoError(t, report.err())
	require.JSONEq(t, `{"claude":{"five_hour":{"utilization":84,"resets_at":"Thu 22 Oct, 01:02","resets_in":"1h 2m","used_dollars":0,"locked_reason":null},"seven_day":null,"error":null},"codex":{"five_hour":null,"seven_day":{"used_percent":20,"resets_at":"Fri 23 Oct, 00:02","resets_in":"1d 2m"},"error":null}}`, mustJSON(t, report))
}

func TestCollectKeepsOtherProviderWhenOneIsRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/claude" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("sensitive-response"))
			return
		}
		_, _ = w.Write([]byte(`{"rate_limit":{"primary_window":{"used_percent":20,"limit_window_seconds":18000,"reset_at":null}}}`))
	}))
	defer server.Close()
	c := client{http: server.Client(), claudeURL: server.URL + "/claude", codexURL: server.URL + "/codex"}
	report := c.collect(context.Background(), stored(login{source: "macOS Keychain"}), stored(login{source: "auth.json"}), time.Now())
	require.NoError(t, report.err())
	require.NotNil(t, report.Claude.Error)
	require.Contains(t, *report.Claude.Error, "401")
	require.Contains(t, *report.Claude.Error, "sign in again using the provider CLI (token read from macOS Keychain)")
	require.NotContains(t, *report.Claude.Error, "sensitive-response")
	require.Nil(t, report.Codex.Error)
	require.Equal(t, float64(20), report.Codex.FiveHour["used_percent"])
}

func TestCollectFailsOnlyWhenEveryProviderFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "300")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()
	c := client{http: server.Client(), claudeURL: server.URL, codexURL: server.URL}
	signedOut := login{err: fmt.Errorf("%w; run `codex login`", errSignedOut)}
	err := c.collect(context.Background(), stored(login{}), stored(signedOut), time.Now()).err()
	require.EqualError(t, err, "no usage to show\nClaude: usage API returned HTTP 429; rate limited; retry in 5m\nCodex: not signed in; run `codex login`")
}

func TestMissingCodexLoginCountsAsSignedOut(t *testing.T) {
	t.Setenv("CODEX_HOME", t.TempDir())
	require.ErrorIs(t, loadCodex().err, errSignedOut)
}

func stored(auth login) func() login {
	return func() login { return auth }
}

func TestCountdownBoundaries(t *testing.T) {
	for _, test := range []struct {
		seconds int
		want    string
	}{
		{-1, "now"}, {0, "now"}, {1, "less than 1m"}, {59, "less than 1m"}, {60, "1m"}, {3600, "1h"}, {90061, "1d 1h 1m"},
	} {
		require.Equal(t, test.want, countdown(time.Duration(test.seconds)*time.Second))
	}
}
