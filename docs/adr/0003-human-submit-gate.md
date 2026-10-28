# 0003. Human submit gate

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §2.7, §3, §6.4, D8

## Context

Submitting an application is irreversible and is sent in the user's name.
Fully unattended submission is an explicit non-goal.

## Decision

- The agent fills the form, takes `handoff/trace/01-filled.png`, posts it with
  the exact text `Form filled. Reply "submit" or "manual".` and **stops**.
- Only the reply `submit` permits clicking the submit/apply button. The reply
  `manual` routes the posting to `manual/`.
- Scheduled (cron) runs may search and save postings but **never** apply.
- `preferences.yml` `apply.require_gate` must stay `true`; `caddie doctor`
  fails if it is `false`.
- After submitting, the confirmation must be verified (URL change or
  confirmation text) and screenshotted (`02-confirmed.png`). An unverifiable
  confirmation routes to `manual/`.

## Consequences

- Every application needs one human message. This is the product's core
  promise, not a limitation (the golf metaphor: the player takes the swing).
- The gate text and rule appear verbatim in `SKILL.md`, and code paths must
  not offer a way to bypass it.
