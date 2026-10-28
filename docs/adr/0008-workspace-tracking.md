# 0008. Tracked workspace dirs; public history, private contact

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §11, §12.2, D14, D15

## Context

The repository is public. The directory layout should come with a clone, but
postings, applications and documents are personal. The career history, on
the other hand, is meant to be public: it doubles as a CV.

## Decision

- `postings/`, `applied/`, `manual/`, `resumes/`, `cover-letters/`,
  `handoff/` and `reviews/` are **tracked, with their contents ignored**.
  Each holds only a `.gitignore` containing `*` and `!.gitignore`. They get
  no README, which would itself be ignored. `AGENTS.md` documents them
  instead.
- `.caddie/`, `.env`, `preferences.yml` and `answers.yml` are ignored
  entirely. Tracked `*.example` files document their shape.
- `history/` is **tracked and public**. Phone, email and anything else not
  for publication live in `preferences.yml` and are merged in at render time
  only.

## Consequences

- `git status` stays clean during normal use.
- Contributors must never commit personal data. Fixtures go in `testdata/`
  and are fictional.
- PDFs and traces are also ignored globally (`*.pdf`, `**/trace/`) as a
  backstop.
