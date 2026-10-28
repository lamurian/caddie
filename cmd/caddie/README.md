# cmd/caddie/

The `caddie` binary: argument parsing, subcommand dispatch and the mapping
from errors to exit codes. The contract (commands, exit codes, `--json`) is
in [`cmd/README.md`](../README.md) and
[ADR 0014](../../docs/adr/0014-cli-contract.md).

## Files

| File | Holds |
|------|-------|
| `main.go` | `env` (stdout, stderr, env lookup, `.env` path), dispatch table, usage |
| `exit.go` | Exit-code constants and the error → code mapping |
| `profile.go` | `caddie profile [--json] [--with-contact]` |
| `doctor.go` | `caddie doctor [--json]`, a thin wrapper over `internal/doctor` |
| `search.go` | `caddie search [--source s] [--limit n] [--include-seen]`; workspace loading helpers |
| `seen.go` | `caddie seen add <id…> --shown\|--rejected` and `caddie seen list [--rejected]` |
| `posting.go` | `caddie posting save <url>` |

Tests run each command in-process against `testdata/` and assert both the
output and the exit code. `profile_test.go` owns `-update` for
`testdata/golden/profile.json`. `runWith` answers HTTP from
`testdata/sources/` so no test touches the network.

## Contributing

- One file per subcommand group; register it in the `commands` map in
  `main.go`. Handlers are plumbing only: load config, call one `internal/`
  function, print, return an error.
- A new error kind needs an exit code in `exit.go`, in the table in
  `cmd/README.md` and in `skills/caddie/SKILL.md` (a test compares them).
- Never read `os.Getenv` or `os.Stdout` directly; go through `env` so tests
  can fake them.
