package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// errSignedOut marks a provider with no stored login, which aiusage leaves out of its display.
var errSignedOut = errors.New("not signed in")

// login is one provider's credentials, or the reason they could not be read.
type login struct {
	token   string
	account string // Codex only
	source  string // where the token came from, named when the provider rejects it
	err     error
}

type credentials struct {
	claude login
	codex  login
}

func loadCredentials(ctx context.Context) credentials {
	home, err := os.UserHomeDir()
	if err != nil {
		err = fmt.Errorf("find home directory: %w", err)
		return credentials{claude: login{err: err}, codex: login{err: err}}
	}
	return credentials{claude: loadClaude(ctx, home), codex: loadCodex(home)}
}

// loadClaude reads the Keychain first on macOS because Claude Code writes logins there.
// A .credentials.json file there can be a stale leftover, so it is only a fallback.
func loadClaude(ctx context.Context, home string) login {
	dir := os.Getenv("CLAUDE_CONFIG_DIR")
	if dir == "" {
		dir = filepath.Join(home, ".claude")
	}
	source := "macOS Keychain"
	data, err := claudeKeychain(ctx, dir)
	if err != nil {
		file := filepath.Join(dir, ".credentials.json")
		fileData, fileErr := os.ReadFile(file)
		switch {
		case fileErr == nil:
			source, data, err = file, fileData, nil
		case !os.IsNotExist(fileErr):
			err = fileErr
		}
	}
	if errors.Is(err, errSignedOut) {
		return login{err: fmt.Errorf("%w; run `claude` to sign in", err)}
	}
	if err != nil {
		return login{err: fmt.Errorf("read credentials: %w; run `claude` to sign in", err)}
	}
	var claude struct {
		OAuth struct {
			AccessToken string `json:"accessToken"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(data, &claude); err != nil {
		return login{err: fmt.Errorf("invalid credentials in %s; run `claude` to sign in", source)}
	}
	if claude.OAuth.AccessToken == "" {
		return login{err: fmt.Errorf("no access token in %s; run `claude` to sign in", source)}
	}
	return login{token: claude.OAuth.AccessToken, source: source}
}

func loadCodex(home string) login {
	dir := os.Getenv("CODEX_HOME")
	if dir == "" {
		dir = filepath.Join(home, ".codex")
	}
	file := filepath.Join(dir, "auth.json")
	data, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return login{err: fmt.Errorf("%w; run `codex login` with file credential storage", errSignedOut)}
	}
	if err != nil {
		return login{err: fmt.Errorf("read credentials: %w; run `codex login` with file credential storage", err)}
	}
	var codex struct {
		Tokens struct {
			AccessToken string `json:"access_token"`
			AccountID   string `json:"account_id"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(data, &codex); err != nil {
		return login{err: fmt.Errorf("invalid credentials in %s; run `codex login` to sign in", file)}
	}
	if codex.Tokens.AccessToken == "" || codex.Tokens.AccountID == "" {
		return login{err: fmt.Errorf("no ChatGPT login in %s; run `codex login` to sign in", file)}
	}
	return login{token: codex.Tokens.AccessToken, account: codex.Tokens.AccountID, source: file}
}
