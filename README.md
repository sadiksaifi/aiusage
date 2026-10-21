# aiusage

See Claude Code and Codex usage together, with local reset dates and time remaining.

```text
Claude
  5-hour: 84% used | resets Thu 22 Oct, 05:19 | resets in 1h 2m

Codex
  7-day: 20% used | resets Tue 27 Oct, 12:13 | resets in 5d 7h 12m
```

## Install

Build with Go 1.27.3 or use the matching binary from `dist/`:

```sh
mise install
mise run install
```

The binary runs on macOS and Linux without Go, jq, curl, or a shell script. macOS Keychain access uses the system-provided `/usr/bin/security` tool. No credentials are copied between machines.

## Usage

```sh
aiusage
aiusage --json
```

The default output omits null fields. Usage percentages are green below 70%, yellow from 70%, and red from 90%. Colors are disabled when piped, when `TERM=dumb`, or when `NO_COLOR` is set.

`--json` keeps null fields and includes `resets_at` and `resets_in` for each available window. Dates use the machine's local timezone. Countdown values look like `5d 7h 12m`, `less than 1m`, or `now`.

## Credentials

Sign in with Claude Code and Codex on each machine before running aiusage. API-key-only logins do not provide subscription usage.

- Claude Code: reads `~/.claude/.credentials.json`, or the Claude Code entry in macOS Keychain when that file is absent. `CLAUDE_CONFIG_DIR` overrides the directory.
- Codex: reads `~/.codex/auth.json`. `CODEX_HOME` overrides the directory. For keyring-only Codex configurations, set `cli_auth_credentials_store = "file"` in the Codex configuration and sign in again.

aiusage reads existing access tokens without refreshing or rewriting them. If a token expires, sign in again with the corresponding provider CLI. Both accounts must be available; failures print an error and return a nonzero exit status.

The app calls the providers' usage endpoints directly. These endpoints can change independently of this project.

## Development

The layout and mise tasks follow `appsum`: a thin `cmd/aiusage` entrypoint and implementation in `internal/app`.

```sh
mise run start
mise run check
mise run build:all
```

Cross-platform binaries are written to `dist/aiusage-{darwin,linux}-{x86_64,arm64}`.
