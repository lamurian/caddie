# 04. Verify, keywords, answers

- **Status:** In progress
- **PRD milestone:** M3 (part 1)

## Goal

The agent can draft a tailored resume and cover letter, prove that every
factual claim traces to `history/` or to stored answers, see how well the
posting's keywords are covered, and run the cover-letter interview without
re-asking questions.

## Inputs

- PRD §6.2, §6.3, §8.3, §8.6, §12.5, §13
- ADRs [0004](../adr/0004-fidelity-check.md),
  [0007](../adr/0007-plain-file-memory.md),
  [0009](../adr/0009-history-schema.md)

## Scope

- `internal/memory/answers`: reads and writes `answers.yml`, with lookup
  order posting → company → global.
  - `caddie answers add --key k --scope global|company:<x>|posting:<slug> "<text>"`
  - `caddie answers show [--for <slug>] [--form]`
  - `caddie answers list`
- `internal/document/verify`: extracts claims (employers, dates, degrees,
  skills, titles and quantified statements) and matches them, exactly or
  fuzzily, against the profile and the stored answers.
  - `caddie verify <doc.md>` exits 1 and lists each untraceable claim.
  - On a pass, it records a content hash in `.caddie/verified/`.
- `internal/document/keywords`: extracts keywords from the posting's
  `## Requirements` and reports which are covered and which are missing.
  - `caddie keywords <posting.md> <doc.md>`.
- Decide and document how non-factual prose is treated (ADR 0004 open
  point).
- Skill: tailor (§6.2) and cover letter (§6.3) workflows, plus the interview
  guidance: at most five questions, batched, using the keys `why_org`,
  `why_role`, `lead_story`, `start_date` and `constraints`.

## Out of scope

PDF rendering (phase 05). Form answers are stored here but used in phase 07.

## Deliverables

`internal/memory/answers*.go`, `internal/document/{verify,keywords}*.go`,
fixtures and golden files, and `skills/caddie/references/{tailor,cover-letter}.md`.

## Definition of Done

- [ ] Verify tests cover exact matches, fuzzy matches (e.g. "Jun 2023" vs
      "June 2023"), false positives (an invented employer or skill caught)
      and answer-sourced claims.
- [ ] Editing a verified file invalidates its hash.
- [ ] The keywords report matches the golden file `testdata/golden/keywords-*.txt`.
- [ ] The answer lookup order is tested across all three scopes;
      `--for <slug>` returns the merged view.
- [ ] `answers.yml` round-trips without losing keys or hand comments
      (or the limitation is documented).
- [ ] The skill tells the agent to run `answers show` before asking anything.
