package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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
	report, err := c.collect(context.Background(), credentials{claudeToken: "claude-token", codexToken: "codex-token", accountID: "account"}, now)
	require.NoError(t, err)
	require.JSONEq(t, `{"claude":{"five_hour":{"utilization":84,"resets_at":"Thu 22 Oct, 01:02","resets_in":"1h 2m","used_dollars":0,"locked_reason":null},"seven_day":null},"codex":{"five_hour":null,"seven_day":{"used_percent":20,"resets_at":"Fri 23 Oct, 00:02","resets_in":"1d 2m"}}}`, mustJSON(t, report))
}

func TestCollectDoesNotExposeHTTPErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("sensitive-response"))
	}))
	defer server.Close()
	c := client{http: server.Client(), claudeURL: server.URL, codexURL: server.URL}
	_, err := c.collect(context.Background(), credentials{}, time.Now())
	require.ErrorContains(t, err, "401")
	require.False(t, strings.Contains(err.Error(), "sensitive-response"))
	require.ErrorContains(t, err, "sign in")
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
