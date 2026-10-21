# aiusage

Claude Code and Codex usage, with local reset times and countdowns.

## Install macOS/Linux

From this checkout:

```sh
CGO_ENABLED=0 GOBIN="$HOME/.local/bin" go install ./cmd/aiusage
```

## Usage

```sh
aiusage
aiusage --json
```

The default output is colored and omits null fields. `--json` includes all fields, with `resets_at` and `resets_in` for each available window.

Sign in to Claude Code and Codex first. Claude credentials come from its credentials file or macOS Keychain; Codex uses `~/.codex/auth.json` with file credential storage. API-key-only logins are unsupported.
