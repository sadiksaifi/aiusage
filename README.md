# aiusage

Claude Code and Codex usage, with local reset times and countdowns.

## Install macOS/Linux

```sh
CGO_ENABLED=0 go install github.com/sadiksaifi/aiusage/cmd/aiusage@latest
```

## Usage

```sh
aiusage
aiusage --json
```

Sign in to Claude Code and Codex first. Claude credentials come from its credentials file or macOS Keychain; Codex uses `~/.codex/auth.json` with file credential storage. API-key-only logins are unsupported.
