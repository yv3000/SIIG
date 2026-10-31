# Changelog

## Unreleased

### Added

- `siig --json` — one machine-readable object on stdout (`version`, `stacks`, `onlineStacks`, `targets`, `scanErrors`) instead of the human summary, for CI jobs that need to report what changed. `file` is always relative to the project, so no local directory layout reaches a CI log; `--json --check` still exits 1 while entries are missing, with the error on stderr. Thanks to @yoreva-01 (#50, #45).

## 2.0.2 — 2026-10-31

### Added

- 12 more embedded stacks, synced from siig-rules: Gleam, Nim, OCaml, Crystal, Fortran, D, PureScript, Lean 4, Coq, LaTeX, PlatformIO and Solidity. The embedded DB is now 122 rules.
- Diagnostics are structured records: `internal/log` wraps `log/slog` and writes WARN/ERROR to stderr, as JSON when `--verbose` is set. An unreadable path and a failed rule-cache read or write now carry the path/stack id that caused them. stdout output is unchanged.

### Fixed

- `--targets docker` on a project with no Docker no longer reports "Everything already up to date"; it lists the target as not applicable and still exits 0. Thanks to @dannyward630 (#51).
- Stacks whose rule was fetched from siig-rules now appear in the detected list as `<Name> (online)`, and in `--verbose` with their source. The online result was computed but never shown, so the documented `Detected: Node.js, Gleam (online)` output could not happen.
- `--offline` applies rules already present in the local cache, as the README states; it used to skip the cache entirely and fall back to embedded rules only.
- The rules sync never picked up a new stack: `git diff` reports nothing for untracked files, which is exactly what an added rule is, so the job declared "already up to date" and opened no pull request. A manual re-run on the same day also failed to push, because the dated branch already existed; each run now gets its own branch.

### Changed

- The project is SIIG. The backronym is retired from the title and the CLI summary.

## 2.0.1 — 2026-10-23

### Added

- `siig --check` — reports what is missing and exits 1, so CI can enforce up-to-date ignore files.
- `siig --clear-cache` — deletes the locally cached remote rules.
- `internal/version` — one version source, stamped from the release tag with `-ldflags`.
- CI: coverage profile with an enforced 80% total gate, Go matrix (1.21 + stable), `govulncheck` audit job, Dependabot for modules and actions.
- Release workflow: version stamping, coverage/vet gates, and a smoke test that runs the built binary (`--version`, `--check`, write-then-check).

### Changed

- `--no-cache` now ignores cached rules for the run instead of deleting `~/.siig/cache`.
- Scanner surfaces paths it could not read instead of silently skipping them.
- Fetched and cached rules are size-capped, parsed with unknown fields rejected, and validated against the requested stack id.
- Target reporting and `--verbose` output are deterministic.
- `.github/workflows/sync-rules.yml` opens a pull request instead of committing and tagging `main` directly.
- golangci-lint moved to v2 configuration; `errcheck` findings fixed rather than excluded.
- Upgraded `cobra` to v1.9.1 and `fatih/color` to v1.18.0.

### Fixed

- A second run detected stacks from the ignore files the first run wrote, breaking the idempotency promise.
- `src/index.js` reported an Express server; any `.vscode` folder reported a VS Code extension; a project directory named after a stack detected that stack.
- `--targets git,gitignore,git` processed the same target repeatedly.
- The writer re-emitted its `# Added by SIIG` header on every append and could glue a block onto a last line that had no newline.
- The online fallback looked at the process working directory rather than the project being populated.
- Unknown duplicate or blank stack ids in `rules/db/` are rejected at load time.
- Release binaries reported a hardcoded version regardless of the tag.

## 2.0.0

- Embedded 110-stack rule DB, online fetch fallback, local cache, and ignore-file writer.
