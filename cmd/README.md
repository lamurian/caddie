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
phase in [`docs/tasks/`](../docs/tasks/README.md) that owns it. Implemented:
`profile [--json] [--with-contact]` and `doctor [--json]` (phase 01);
`search [--source s] [--limit n] [--include-seen]`,
`seen add <id…> --shown|--rejected`, `seen list [--rejected]` and
`posting save <url>` (phase 02); `search --source portal[:<org>]` and `posting save` for portal URLs
(phase 03); `answers add --key k --scope s "<text>"`,
`answers show [--for <slug>] [--form]` and `answers list` (phase 04,
[ADR 0020](../docs/adr/0020-answers-scopes-and-form-keys.md));
`verify <doc.md> [--json]` and `keywords <posting.md> <doc.md> [--json]`
(phase 04, [ADR 0021](../docs/adr/0021-claim-extraction-and-non-factual-prose.md));
`render resume|cover-letter <slug> [--json]` (phase 05,
[ADR 0022](../docs/adr/0022-render-naming-and-gate.md)). Flags may
come before or after arguments. A malformed `answers.yml` is exit 4; an
unknown `--for` slug or a missing file is exit 5.

## Exit codes

Defined in `cmd/caddie/exit.go`, the contract of
[ADR 0014](../docs/adr/0014-cli-contract.md). Errors go to stderr as one line,
`caddie <command>: <message>`.

<!-- exit-codes:begin -->
| Code | Meaning | What the agent does |
|------|---------|---------------------|
| 0 | Success | Continue |
| 1 | A check failed (`doctor` found a problem; `verify` found an untraceable claim) | Read the output, fix the cause, rerun |
| 2 | Usage error: unknown command, bad flag or argument | Fix the command line |
| 3 | Configuration error: `.env`, an env var or `preferences.yml` is invalid or missing | Show the message to the user |
| 4 | Malformed input: `history/` or a posting breaks its schema; the message names file and line | Show the file and line to the user |
| 5 | Not found: no job posting at a URL, an unknown slug or a missing file | Check the URL, slug or path |
| 6 | Illegal posting status transition | Do not edit `status` by hand; follow the workflow |
| 7 | Refused by a safety rule: `render` on an unverified or edited draft (reserved: upload outside `/handoff`) | Stop; never work around it |
| 8 | External tool or network failure: HTTP, `quarto` (later: `docker`) | Run `caddie doctor`, retry once, then tell the user |
| 9 | Internal error | Report it to the user |
<!-- exit-codes:end -->

## Conventions

- Every command supports `--json`. Keep output short and structured, because
  the agent reasons over exactly what is printed.
- Exit codes are documented in the table above, mirrored verbatim in
  `skills/caddie/SKILL.md` (a test compares them). Change both together.
- One file per subcommand group (e.g. `cmd/caddie/answers.go`) so that
  `main.go` stays small; every file stays under 300 lines.
- Standard library `flag` (or a single small dependency) is preferred; adding
  a CLI framework needs an ADR.

## Contributing

1. Add or change the handler in its own file.
2. Add a test that runs the command against `testdata/`, and assert both the
   output and the exit code.
3. Update the CLI table in `skills/caddie/SKILL.md` if the surface changed.
