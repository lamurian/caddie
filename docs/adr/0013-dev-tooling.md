# 0013. Dev tooling: mise, lefthook and Renovate

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §11 (layout), §13 (testing)

## Context

Contributors are mostly coding agents in fresh containers. They need the
same tool versions everywhere, the same checks locally and in CI, and
dependencies that stay current without manual chores. PRD §11 lists a
`Makefile`, but `make` pins no tool versions.

## Decision

- **mise** (`mise.toml`) pins every dev and runtime tool at an exact version
  (Go, golangci-lint, lefthook, gitleaks, shellcheck, actionlint, yamllint,
  Quarto) and records checksums in the committed `mise.lock`.
- **mise tasks replace the Makefile.** `mise run check` is the single entry
  point that CI, the pre-push hook and contributors all run. Go tasks
  (`build`, `test`, `vet`) join it in phase 01.
- **Complexity gate:** `.golangci.yml` enables gocyclo and gocognit at
  `min-complexity: 15`, so every Go function, tests included, keeps
  cyclomatic and cognitive complexity ≤ 15. A failing function is split, not
  silenced with `//nolint`. `lint:go` skips until `go.mod` exists.
- mise does **not** load `.env`; `internal/config` owns the precedence
  env > `.env` > defaults. `GOTOOLCHAIN=local` keeps mise's Go authoritative.
- **lefthook** (`lefthook.yml`) runs fast, staged-file checks on pre-commit:
  gitleaks on staged changes, a workspace-directory guard
  (`scripts/guard-workspace.sh`), gofmt, shellcheck, yamllint, actionlint and
  the line limits. pre-push runs `mise run check`. `mise install` installs the
  hooks.
- **Renovate** (`.github/renovate.json5`) updates mise tools, Go modules,
  GitHub Actions (pinned by digest) and, from phase 06, compose images. It
  waits 3 days after a release, runs weekly, groups the Go toolchain
  (mise + `go.mod`) into one PR, and **automerges every update, majors
  included, once the required `check` CI job passes**. Vulnerability fixes
  skip the schedule and the waiting period.

## Consequences

- One command reproduces CI locally; hooks never drift from CI.
- Automerge is only as safe as CI: `main` must require the `check` job, and
  phase work must keep test coverage meaningful (PRD §13).
- A major bump that passes CI but changes behaviour can land unreviewed; the
  Dependency Dashboard and git history are the audit trail.
- Hooks can be bypassed with `--no-verify`; CI remains authoritative.
- Contributors need mise installed; everything else follows from it.
