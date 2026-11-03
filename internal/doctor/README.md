# internal/doctor/

The checks behind `caddie doctor`. `Run(Options)` returns a `Report` of
`ok` / `warn` / `fail` lines; it reports and never changes anything.

| Check | Fails when |
|-------|-----------|
| `config` | `.env` or an env var is invalid (later checks are skipped) |
| `workspace` | a workspace directory is missing |
| `trace` | `<state>/trace` is missing (`local` mode: `warn`) or not a directory |
| `preferences` | `preferences.yml` is missing or malformed |
| `gate` | `apply.require_gate` is not `true` ([ADR 0003](../../docs/adr/0003-human-submit-gate.md)) |
| `history` | `history/` is malformed (message names file and line) |
| `quarto` | `quarto` is found but cannot render a test PDF (missing: `warn`) |
| `docker` | never: a missing `docker` is a warning (stack checks skipped) |
| `browser-stack` | a compose service is not running (`local` mode: `warn`, all skipped) |
| `ports` | a service declares `ports:` or a container publishes a host port |
| `proxy-deny` | the proxy does not answer CONNECT `403` for `example.com` |
| `proxy-allow` | the proxy does not answer CONNECT `200` for `careers.who.int` |

The `quarto` check renders a one-line `doctor.qmd` (`format: typst`) to
`doctor.pdf` in a fresh `os.MkdirTemp` dir outside the workspace, with a
2-minute timeout, and removes the dir afterwards. A failure shows the last
output lines, since `caddie render` would fail the same way.

The stack checks (`browser.go`, [ADR 0023](../../docs/adr/0023-doctor-browser-stack.md))
run `docker compose -f <CADDIE_COMPOSE_FILE> config|ps|exec` through
`browse.Exec`, 45 s per call. They run only in `compose` mode with `docker`
on PATH; the proxy probes curl from the browser container and only run when
every service is up.

The `trace` check ([ADR 0024](../../docs/adr/0024-move-archive-and-trace.md))
stats `<CADDIE_STATE_DIR>/trace`, the source of the browser's `/trace`
mount. Writability for uid 1000 is only guessed, on Linux, from the mode
bits: a dir owned by another uid with no group/other write `warn`s (an ACL
for uid 1000 sets the group mask, so it passes). The compose file mounts
`../.caddie/trace` whatever `CADDIE_STATE_DIR` says.

Only `fail` makes the report unhealthy; the CLI then exits 1.

## Contributing

- One small function per check; inject anything external through
  `Options` (`Lookup`, `LookPath`, `Quarto`, `Browser`) so tests stay offline.
- A new phase adds its own checks here (e.g. templates, browser stack) with
  a test for the healthy and the failing case.
- Messages say what is wrong and how to fix it, in one line.
