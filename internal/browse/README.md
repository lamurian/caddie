# internal/browse/

Runs agent-browser (Vercel Labs) for `caddie browse <session> <args…>`, the
only way the agent drives the sandboxed browser
([PRD §7, §9](../../PROPOSAL.md),
[ADR 0002](../../docs/adr/0002-sandboxed-browser.md),
[ADR 0005](../../docs/adr/0005-credentials.md)). Depends only on `config`.

| File | Holds |
|------|-------|
| `browse.go` | Sentinel errors, `Exec` interface, `OSExec` (os/exec), `Browse` |
| `argv.go` | `Argv`: the compose or local command line |
| `validate.go` | `ValidateSession`, `ValidateArgs`, `ValidateUploadPath`, `HandoffDir` |
| `trace.go` | `ValidateScreenshotPath`, `TraceDir` |
| `refuse.go` | The `batch` refusal and `refusedFlags` |

## Browse

`Browse(ctx, cfg, exec, session, args, stdout, stderr)` validates, builds
the argv, wraps `ctx` in `context.WithTimeout(cfg.BrowseTimeout)` and runs
it, streaming output. Nothing runs unless validation passes.

Argv, with the session as agent-browser's global `--session <name>` flag
and the caller's args appended verbatim:

```
compose: docker compose -f <workspace>/<ComposeFile> exec -T <ComposeService> agent-browser --session <session> <args…>
local:   <BrowserBin> --session <session> <args…>
```

`-T` disables TTY allocation so output pipes cleanly. Any other
`CADDIE_BROWSER_MODE` wraps `config.ErrConfig`.

## What is validated

- **Session name** (`ValidateSession`): hostname-like, i.e. dot-separated
  labels of 1–63 characters from `a-z 0-9 -`, no label starting or ending
  with `-`, no empty label (so no leading, trailing or double dot), at most
  253 characters. **Uppercase is rejected**, not folded, so one portal maps
  to one session and one state file in the browser volume.
- **No session override** (`ValidateArgs`): `--session x` or `--session=x`
  in the args is rejected, since it would bypass the check above.
- **No `batch`** (`ValidateArgs`, `refuse.go`): a bare `batch` token is
  refused (`ErrBatchRefused`). Its commands arrive as one quoted token or
  on stdin, so the upload and screenshot checks could not see them.
- **No sandbox overrides** (`refuse.go`): `--profile`, `--state`,
  `--session-name`, `--cdp`, `--auto-connect`, `--executable-path`,
  `--extension`, `--args`, `--proxy`, `--proxy-bypass`, `--provider`,
  `--config` and `--allow-file-access`, bare or `=value`, anywhere in the
  args (`ErrFlagRefused`).
- **Uploads** (`ValidateArgs`, `ValidateUploadPath`): at the first bare
  `upload` token, `upload <selector> <file…>` needs at least one file and
  every argument after the selector must, after `path.Clean`, be absolute
  and strictly inside `/handoff` (the container's read-only mount of
  `handoff/`). Rejected: relative paths, `/handoff` itself,
  `/handoffx/a.pdf`, `/handoff/../etc/passwd`. A bare `upload` token used
  as a value also triggers the check; that is deliberate. `/trace/…` is
  refused too: uploads come from `/handoff` only.
- **Screenshots** (`ValidateArgs`, `ValidateScreenshotPath`, `trace.go`):
  at the first bare `screenshot` token, every positional after it must be
  strictly inside `/trace` after `path.Clean` (the read-write mount of
  `.caddie/trace/`, [ADR 0024](../../docs/adr/0024-move-archive-and-trace.md)),
  and a `--screenshot-dir` value must be `/trace` or inside it. No path at
  all is allowed (agent-browser's temporary default). Flags are skipped;
  `--threshold`, `--screenshot-format` and `--screenshot-quality` take a
  value (agent-browser 0.39.2); any other flag is assumed to take none, so
  an unknown flag's value is checked as a path and refused. Rejected:
  `/handoff/x.png`, `trace/x.png`, `/trace/../etc/x.png`, `/tracex/a.png`.

Not validated: `pdf`, `download`, `state save|load` and `eval`. They act
inside the container, whose host mounts are the read-only `/handoff` and
`/trace`, so at worst they write into `.caddie/trace/`. Page and posting
content stays data; nothing here interprets it. `eval` can click like
`click` can; the submit gate is a workflow rule, not an argv check.

## Errors

| Error | Meaning | Exit (CLI) |
|-------|---------|------------|
| `ErrInvalidSession` | bad session name or `--session` in args | 7 |
| `ErrUploadOutsideHandoff` | an upload path outside `/handoff` | 7 |
| `ErrScreenshotOutsideTrace` | a screenshot path outside `/trace` | 7 |
| `ErrBatchRefused` | the `batch` command | 7 |
| `ErrFlagRefused` | a flag that overrides the sandbox (`refuse.go`) | 7 |
| `ErrExec` | agent-browser could not start or exited non-zero | 8 |
| `ErrTimeout` | `CADDIE_BROWSE_TIMEOUT` expired; always wrapped with `ErrExec` | 8 |

## Contributing

- Every new check gets a table-driven case in `validate_test.go`, and must
  reject before `Exec` runs (see `TestBrowseRejectsBeforeExec`).
- Tests use the fake `Exec` in `browse_test.go`; they never start Docker,
  a browser or the network.
- Never loosen a check to make a workflow pass; extend `handoff/` or the
  proxy allowlist instead ([ADR 0002](../../docs/adr/0002-sandboxed-browser.md)).
