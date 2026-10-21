package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type credentials struct {
	claudeToken string
	codexToken  string
	accountID   string
}

func loadCredentials(ctx context.Context) (credentials, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return credentials{}, fmt.Errorf("find home directory: %w", err)
	}
	claudeDir := os.Getenv("CLAUDE_CONFIG_DIR")
	if claudeDir == "" {
		claudeDir = filepath.Join(home, ".claude")
	}
	data, err := os.ReadFile(filepath.Join(claudeDir, ".credentials.json"))
	if err != nil && os.IsNotExist(err) && runtime.GOOS == "darwin" {
		data, err = claudeKeychain(ctx, claudeDir)
	}
	if err != nil {
		return credentials{}, fmt.Errorf("read Claude credentials: %w; run `claude` to sign in", err)
	}
	var claude struct {
		OAuth struct {
			AccessToken string `json:"accessToken"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(data, &claude); err != nil {
		return credentials{}, fmt.Errorf("invalid Claude credentials; run `claude` to sign in")
	}
	if claude.OAuth.AccessToken == "" {
		return credentials{}, fmt.Errorf("no Claude access token; run `claude` to sign in")
	}
	codexDir := os.Getenv("CODEX_HOME")
	if codexDir == "" {
		codexDir = filepath.Join(home, ".codex")
	}
	data, err = os.ReadFile(filepath.Join(codexDir, "auth.json"))
	if err != nil {
		return credentials{}, fmt.Errorf("read Codex credentials: %w; run `codex login` with file credential storage", err)
	}
	var codex struct {
		Tokens struct {
			AccessToken string `json:"access_token"`
			AccountID   string `json:"account_id"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(data, &codex); err != nil {
		return credentials{}, fmt.Errorf("invalid Codex credentials; run `codex login` to sign in")
	}
	if codex.Tokens.AccessToken == "" || codex.Tokens.AccountID == "" {
		return credentials{}, fmt.Errorf("no Codex ChatGPT login; run `codex login` to sign in")
	}
	return credentials{claudeToken: claude.OAuth.AccessToken, codexToken: codex.Tokens.AccessToken, accountID: codex.Tokens.AccountID}, nil
}
