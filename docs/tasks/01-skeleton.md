# 01. Skeleton

- **Status:** Done
- **PRD milestone:** M1

## Goal

A buildable `caddie` binary that loads configuration, understands the
workspace layout and parses `history/` into a JSON profile. `caddie doctor`
reports the health of the setup, and CI builds and tests every push.

## Inputs

- PRD §7 (CLI), §11 (layout), §12.1–12.2 (history), §12.7–12.8 (config), §13
- ADRs [0001](../adr/0001-skill-plus-cli.md),
  [0008](../adr/0008-workspace-tracking.md),
  [0009](../adr/0009-history-schema.md),
  [0011](../adr/0011-posting-state-machine.md),
  [0012](../adr/0012-agentic-docs-and-line-limits.md)

## Scope

- `go.mod` (its `go` directive matching `mise.toml`), mise tasks `build`,
  `test` and `vet` joined into `check` (`lint:go` already exists)
  ([ADR 0013](../adr/0013-dev-tooling.md)), and
  `cmd/caddie/main.go` with subcommand dispatch, `--json` and an exit-code
  table.
- `internal/config`: `.env` plus env vars (env wins) and `preferences.yml`.
  Defaults match `.env.example`.
- `internal/workspace`: directory layout, slug generation, the posting
  frontmatter type and the state machine (transitions only; no moves yet).
- `internal/profile`: history parser and `about.md` frontmatter → struct →
  JSON.
- `caddie profile`: JSON summary of the history plus preferences. Contact
  details are not included unless `--with-contact` is passed.
- `caddie doctor` checks: the workspace dirs exist, config parses,
  `history/` parses, and `apply.require_gate` is `true`. Missing `quarto` and
  `docker` are reported as warnings for now.
- `testdata/history/`: a fictional CV covering the schema's edge cases.
- CI: nothing to add; the `check` job already runs `mise run check`.

## Out of scope

Sources, memory, documents, browser, `move`.

## Deliverables

`go.mod`, Go tasks in `mise.toml`, `cmd/caddie/`,
`internal/{config,workspace,profile,doctor}/` (`doctor` keeps the checks out
of `cmd/`), `testdata/history/`, `testdata/golden/profile.json`.

## Definition of Done

- [x] `mise run check` runs build, vet, tests, golangci-lint (complexity
      ≤ 15) and the existing checks, and passes locally and in CI.
- [x] Env vars override `.env`, which overrides defaults (tested).
- [x] The history parser handles title, organisation, location, date range,
      `Present` and bullets; publications and presentations parse; errors
      name the file and line.
- [x] `caddie profile --json` matches `testdata/golden/profile.json`.
- [x] Profile output omits phone and email by default.
- [x] State machine tests cover every legal transition and reject an
      illegal one.
- [x] `caddie doctor` exits non-zero with a clear message when
      `require_gate: false` or `history/` is malformed.
- [x] The exit-code table is documented in `cmd/README.md`.
- [x] `skills/caddie/SKILL.md` exists, with what Caddie is, the directory
      layout and the hard rules (PRD §8.1, §8.3).
