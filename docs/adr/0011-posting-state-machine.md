# 0011. Posting frontmatter as a state machine; MIT licence

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §6.5, §12.3, D19, D20

## Context

Each posting moves through a lifecycle, and both agents must see and update
it the same way. The repository is public and has no bundled dependency on
Paseo.

## Decision

- Postings are `postings/<slug>.md` with **YAML frontmatter**: slug, title,
  company, url, source, ats, ats_host, location, work_mode, salary
  min/max/currency/period, deadline, status, saved_at, applied_at,
  follow_up_after and manual_reason.
- `status` is a state machine owned by `internal/workspace`:
  `saved → tailored → rendered → applied`, and any pre-applied state →
  `manual`. Only `caddie` commands change it; illegal transitions are
  errors.
- `caddie move <slug> applied|manual` relocates the posting, resume (md +
  pdf), cover letter (md + pdf) and trace together. It also updates the
  frontmatter and clears `handoff/`.
- Licence: **MIT**. Paseo's skills are installed by the user and are not part
  of this repository.

## Consequences

- State can be checked by code and by `grep`. `follow-ups` computes overdue
  items and approaching deadlines from frontmatter alone.
- Hand edits to `status` are possible but are validated on the next command.
