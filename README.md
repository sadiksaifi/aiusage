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

Sign in to Claude Code, Codex, or both first. On macOS, Claude credentials come from the Keychain, with `~/.claude/.credentials.json` as a fallback; on Linux, from that file. Codex uses `~/.codex/auth.json` with file credential storage. API-key-only logins are unsupported.

aiusage shows every provider you are signed in to and skips the others. When one provider fails, its error appears in place of its usage, and the other provider still shows. `--json` reports that error in the provider's `error` field. aiusage exits with an error only when no provider has usage to show.
