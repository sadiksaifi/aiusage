# aiusage

## Sources of truth

- [README.md](README.md): user-facing behavior. Keep edits scoped, examples generic, and screenshots real.
- [.mise.toml](.mise.toml): tools, tasks, and setup.
- For release, installer, or upgrade changes, read [.tagsmith.jsonc](.tagsmith.jsonc), [release.yml](.github/workflows/release.yml), [.goreleaser.yml](.goreleaser.yml), [cliff.toml](cliff.toml), [install.sh](install.sh), and [upgrade.go](internal/app/upgrade.go) together.

Document only what code and configuration cannot express. State each fact once and point to its owner elsewhere.

## Design

- Minimize owned code. Prefer maintained libraries and Charm's default styles. Call provider usage endpoints directly rather than through usage-fetching libraries.
- Build deep modules with narrow interfaces. Inject dependencies at replaceable boundaries; add abstractions for a concrete need.
- Keep tests few and behavioral. Cover owned risks, not library behavior or wiring. Keep test-only seams out of production code.

## Work

- Use mise tasks and mise-managed tools. Verify changed behavior; distinguish native execution from cross-compilation.
- Preview release tags with [Tagsmith's](https://tagsmith.site) `--dry-run --json` before an authorized mutation. Existing release tags stay frozen.
