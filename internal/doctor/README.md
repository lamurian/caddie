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
| `quarto` | `quarto` is found but cannot render a test PDF (missing: `warn`) |
| `docker` | never: a missing `docker` is a warning for now |

The `quarto` check renders a one-line `doctor.qmd` (`format: typst`) to
`doctor.pdf` in a fresh `os.MkdirTemp` dir outside the workspace, with a
2-minute timeout, and removes the dir afterwards. A failure shows the last
output lines, since `caddie render` would fail the same way.

Only `fail` makes the report unhealthy; the CLI then exits 1.

## Contributing

- One small function per check; inject anything external through
  `Options` (`Lookup`, `LookPath`, `Quarto`) so tests stay offline.
- A new phase adds its own checks here (e.g. templates, browser stack) with
  a test for the healthy and the failing case.
- Messages say what is wrong and how to fix it, in one line.
