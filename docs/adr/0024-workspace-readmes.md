# 0024. Workspace directories carry a tracked README

- **Status:** Accepted
- **Date:** 2026-11-03
- **PRD refs:** §11, D14
- **Supersedes:** the "no README" clause of [0008](0008-workspace-tracking.md)

## Context

ADR 0008 kept workspace directories to a lone `.gitignore` and documented
them only in `AGENTS.md`. Readers who open `handoff/`, `manual/` or
`reviews/` find nothing explaining what belongs there, and AGENTS.md rule 6
asks every directory for a README.

## Decision

- Each of `postings/`, `applied/`, `manual/`, `resumes/`, `cover-letters/`,
  `handoff/` and `reviews/` tracks a `README.md` next to its `.gitignore`.
- Their `.gitignore` is `*`, `!.gitignore`, `!README.md`.
- `scripts/guard-workspace.sh` allows exactly `<dir>/.gitignore` and
  `<dir>/README.md` at the top level; anything else, and anything in
  `.caddie/`, is rejected.
- The rest of ADR 0008 stands unchanged.

## Consequences

- Each directory explains itself where the user looks.
- The readmes must not describe personal content by example; fictional
  examples only.
- A user file named `README.md` in a workspace directory would be tracked.
  The name is reserved there.
