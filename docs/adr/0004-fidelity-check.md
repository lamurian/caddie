# 0004. Mandatory fidelity check before rendering

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §6.2, §6.3, §7, §13, D9

## Context

LLMs embellish when tailoring: they invent skills, stretch dates and upgrade
titles. A resume or cover letter with an untraceable claim is a liability for
the user.

## Decision

- `caddie verify <doc.md>` checks every employer, date, degree, skill and
  claim against `history/` or a stored answer in `answers.yml`. It exits 1
  on any untraceable claim and lists each one.
- On success it records a content hash in `.caddie/verified/`.
- `caddie render` **refuses** unless the exact content it is about to render
  has a matching verified hash.
- `caddie keywords` reports coverage of the posting's requirements as a
  separate, advisory report. Keywords are covered only where the history
  supports them.

## Consequences

- Verification is deterministic code with tests for exact, fuzzy,
  false-positive and stored-answer cases. Matching heuristics must be
  conservative and well tested.
- Any edit after verifying requires verifying again. That is intended.
- Prose such as "passionate about…" needs a rule: either it is classified as
  non-factual or it must trace to an answer. This is decided in phase 04.
