# 0024. Move: archive layout, rollback, trace mount and manual steps

- **Status:** Accepted
- **Date:** 2026-11-03
- **PRD refs:** §6.4, §6.5, §12.3, D8, D11, D12

## Context

Phase 07 adds `caddie move` and the manual checklist. ADR 0011 says what
moves, not where it lands, how a failure is undone, or how trace
screenshots get written: `/handoff` is read-only in the browser
container (ADR 0002), so PRD §6.4's `screenshot handoff/trace/…` cannot
work.

## Decision

- **Trace mount.** `docker/compose.yml` bind-mounts the host's
  `.caddie/trace/` read-write at `/trace`. The agent runs
  `caddie browse <host> screenshot /trace/01-filled.png` and later
  `/trace/02-confirmed.png`. One application is in flight at a time, like
  `handoff/`. `caddie browse` refuses a `screenshot` path outside `/trace`
  (exit 7); uploads stay limited to `/handoff`, never `/trace`.
- **Archive layout.** `caddie move <slug> applied|manual` writes
  `<dest>/<slug>.md` (so every `*.md` glob over `applied/` and `manual/`
  still finds postings) and a sibling directory `<dest>/<slug>/` holding
  whatever exists of `resume.md`, the resume PDF, `cover-letter.md`, the
  cover-letter PDF (both named by `document.PDFName`) and `trace/` (the
  files of `.caddie/trace/`). Missing artefacts are skipped and listed.
- **Frontmatter.** `applied`: `applied_at` = today, `follow_up_after` =
  today + `apply.follow_up_days`. `manual`: `manual_reason` = `--reason`
  (required, one line), dates left as they are. Transitions follow
  ADR 0011 (`applied` only from `rendered`).
- **Preconditions, checked before anything moves.** The posting is in
  `postings/`; the transition is legal (exit 6); neither `<dest>/<slug>.md`
  nor `<dest>/<slug>/` exists (`workspace.ErrDestExists`, exit 6).
- **Rollback.** Moves are `os.Rename` (copy + remove across devices).
  Every completed step is recorded; on the first failure they are undone
  in reverse, the created directory is removed, and the error names any
  step that could not be undone. The posting is written last, atomically,
  and the source posting removed after it, so a half-moved state never
  shows the posting in two places.
- **Handoff and trace.** After a successful move the regular files in
  `handoff/` (except `.gitignore`) are deleted and `.caddie/trace/` is
  left empty but present (it is a bind-mount source). A failure there is
  reported (exit 9) but does not undo the move.
- **Manual steps.** `caddie posting steps <slug> "<step>"…` writes or
  updates the posting's `## Manual steps` section of `- [ ]` items
  (`posting.SetManualSteps`). Existing items keep their order and tick
  state; a step already present (trimmed, case-insensitive) is not added
  twice; the rest of the file is kept byte for byte.
- **Output.** `caddie move` prints JSON: `slug`, `status`, `posting`,
  `moved`, `skipped`, `applied_at`, `follow_up_after`, `manual_reason`.

## Consequences

- No code path clicks submit: `move` never calls the browser.
- `.caddie/trace/` must exist before `docker compose up`; doctor and the
  docker README say how. Its owner must let the browser's uid 1000 write.
- PRD §6.4's `handoff/trace/` path is superseded by `/trace`.
- Follow-ups (phase 08) read `applied/*.md` and `manual/*.md` unchanged.
