# cmd/

The `caddie` CLI entry point lives in `cmd/caddie/main.go`
([PRD §7](../PROPOSAL.md#7-cli-specification)).

## What belongs here

- Argument parsing and subcommand dispatch only.
- Wiring: load config, build the workspace, call into `internal/` packages,
  print results, map errors to exit codes.
- **No business logic.** If a subcommand handler grows beyond plumbing, move
  the logic into the relevant `internal/` package where it can be unit-tested.

## Subcommands

`profile`, `search`, `seen`, `posting save`, `answers`, `verify`, `keywords`,
`render`, `browse`, `move`, `follow-ups`, `doctor`. Each is introduced by the
phase in [`docs/tasks/`](../docs/tasks/README.md) that owns it.

## Conventions

- Every command supports `--json`. Keep output short and structured, because
  the agent reasons over exactly what is printed.
- Exit codes are documented in one table, mirrored verbatim in
  `skills/caddie/SKILL.md`. Change both together.
- One file per subcommand group (e.g. `cmd/caddie/answers.go`) so that
  `main.go` stays small; every file stays under 300 lines.
- Standard library `flag` (or a single small dependency) is preferred; adding
  a CLI framework needs an ADR.

## Contributing

1. Add or change the handler in its own file.
2. Add a test that runs the command against `testdata/`, and assert both the
   output and the exit code.
3. Update the CLI table in `skills/caddie/SKILL.md` if the surface changed.
