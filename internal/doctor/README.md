# internal/doctor/

The checks behind `caddie doctor`. `Run(Options)` returns a `Report` of
`ok` / `warn` / `fail` lines; it reports and never changes anything.

| Check | Fails when |
|-------|-----------|
| `config` | `.env` or an env var is invalid (later checks are skipped) |
| `workspace` | a workspace directory is missing |
| `preferences` | `preferences.yml` is missing or malformed |
| `gate` | `apply.require_gate` is not `true` ([ADR 0003](../../docs/adr/0003-human-submit-gate.md)) |
| `history` | `history/` is malformed (message names file and line) |
| `quarto`, `docker` | never: missing tools are warnings for now |

Only `fail` makes the report unhealthy; the CLI then exits 1.

## Contributing

- One small function per check; inject anything external through
  `Options` (`Lookup`, `LookPath`) so tests stay offline.
- A new phase adds its own checks here (e.g. templates, browser stack) with
  a test for the healthy and the failing case.
- Messages say what is wrong and how to fix it, in one line.
