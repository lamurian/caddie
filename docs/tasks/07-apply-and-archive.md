# 07. Apply and archive

- **Status:** Todo
- **PRD milestone:** M4 (part 2)

## Goal

"apply <slug>" fills the form in the sandbox, stops at the human gate, then
either submits and archives with a trace, or writes a manual checklist and
routes the posting to `manual/`.

## Inputs

- PRD §6.4, §6.5, §8.3–8.5, §12.3
- ADRs [0003](../adr/0003-human-submit-gate.md),
  [0005](../adr/0005-credentials.md),
  [0006](../adr/0006-portals-and-manual.md),
  [0011](../adr/0011-posting-state-machine.md)

## Scope

- `internal/workspace.Move`: `caddie move <slug> applied|manual [--reason …]`.
  - Moves the posting, resume (md + pdf), cover letter (md + pdf) and trace
    together, as atomically as the filesystem allows.
  - Updates `status`, `applied_at`, `follow_up_after` (from
    `apply.follow_up_days`) and `manual_reason`.
  - Clears `handoff/`.
- `internal/posting`: a helper that writes or updates the `## Manual steps`
  checklist (`- [ ]` items) in a posting.
- Form answers: `caddie answers show --for <slug> --form` drives the form;
  missing answers are asked once and stored.
- Skill (`references/apply.md`):
  - browser steps (open, `snapshot -i`, fill, upload, select, screenshot);
  - the gate text, verbatim;
  - the submit-confirmation check;
  - the manual triggers (login wall, CAPTCHA, unsupported portal,
    unverifiable confirmation);
  - the credential rules.
- Exit-code table completed in `references/exit-codes.md`.

## Out of scope

`follow-ups` reporting (phase 08).

## Deliverables

`internal/workspace/move*.go`, `internal/posting/manual*.go`, the CLI
handler, and `skills/caddie/references/{apply,exit-codes}.md`.

## Definition of Done

- [ ] `Move` tests: all artefacts move together, frontmatter is updated,
      `handoff/` is cleared, and missing optional artefacts are tolerated.
- [ ] A failed move leaves no half-moved state (or recovery is documented
      and tested).
- [ ] The manual checklist helper is idempotent and preserves existing
      ticked items.
- [ ] The skill contains the gate text and all six hard rules verbatim
      (PRD §8.3).
- [ ] A dry run against a public test form (Greenhouse demo or similar)
      reaches the gate and stops. Documented in the PR.
- [ ] No code path clicks submit; only the agent does, after `submit`.
