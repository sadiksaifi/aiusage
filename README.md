# aiusage

Claude Code and Codex usage, with local reset times and countdowns.

## Install macOS/Linux

```sh
curl -fsSL https://github.com/sadiksaifi/aiusage/releases/latest/download/install.sh | sh
```

Installs to `~/.local/bin`.

## Usage

```sh
aiusage
aiusage --json
aiusage --upgrade
aiusage -v
```

Sign in to Claude Code and Codex first. Claude credentials come from its credentials file or macOS Keychain; Codex uses `~/.codex/auth.json` with file credential storage. API-key-only logins are unsupported.
