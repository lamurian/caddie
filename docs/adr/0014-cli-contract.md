# 0014. CLI contract: exit codes, output and configuration resolution

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §7, §8.4, §12.7–12.8, D2

## Context

The agent reasons over exactly what `caddie` prints and over its exit code
(ADR 0001). PRD §7 asks for documented non-zero exit codes and `--json`
everywhere, but does not fix the codes, where `.env` is read from, how
relative paths resolve or what an empty variable means.

## Decision

- **Exit codes** are fixed by category, defined in `cmd/caddie/exit.go` and
  documented in one table in `cmd/README.md`, which `skills/caddie/SKILL.md`
  mirrors verbatim (a test compares them):
  0 ok, 1 check failed, 2 usage, 3 config, 4 malformed input, 5 not found,
  6 illegal status transition, 7 refused by a safety rule, 8 external tool
  or network failure, 9 internal error. Codes are never renumbered; a new
  category takes the next free number.
- Packages expose **sentinel errors** (`config.ErrConfig`,
  `profile.ErrMalformed`, `workspace.ErrIllegalTransition`…); `cmd` maps
  them to codes with `errors.Is`. Anything unmapped is 9.
- **Output:** results go to stdout; errors go to stderr as one line,
  `caddie <command>: <message>`. Data commands (`profile`) always print JSON;
  report commands (`doctor`) print short text unless `--json` is given.
- **Configuration:** `.env` is read from the current directory. Precedence
  is env > `.env` > built-in defaults (equal to `.env.example`, tested). A
  variable set but empty falls back to its default. Unknown `CADDIE_*` keys
  in `.env` and invalid enum or duration values are config errors.
- **Paths:** `CADDIE_WORKSPACE` is relative to the current directory; every
  other path setting is relative to the workspace.
- `preferences.yml` is decoded strictly (unknown keys are errors) on top of
  defaults; `apply.require_gate` defaults to `true` when omitted.
- YAML is parsed with `go.yaml.in/yaml/v3`, the maintained successor of
  `gopkg.in/yaml.v3`; slugs use `golang.org/x/text` to strip diacritics.

## Consequences

- Agents can branch on the exit code alone, e.g. 4 means "show the user
  file:line", 7 means "stop".
- A typo in `.env` or `preferences.yml` fails loudly instead of silently
  using a default. This includes a mistyped `require_gate`.
- Changing a code or the table means editing `exit.go`, `cmd/README.md` and
  `SKILL.md` together.
