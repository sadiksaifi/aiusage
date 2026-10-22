# aiusage

Claude Code and Codex usage, with local reset times and countdowns.

## Install macOS/Linux

### Homebrew

```sh
brew install sadiksaifi/tap/aiusage
```

### curl

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

`--upgrade` uses the owning Homebrew installation or, for curl installations, a SHA-256-verified release download. Homebrew failures never fall back to replacing the binary directly.

Sign in to Claude Code and Codex first. Claude credentials come from its credentials file or macOS Keychain; Codex uses `~/.codex/auth.json` with file credential storage. API-key-only logins are unsupported.
