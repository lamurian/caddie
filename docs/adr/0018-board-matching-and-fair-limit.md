# 0018. Company boards match titles; --limit is shared across feeds

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §6.1, §7

## Context

ADR 0016 kept a job when a `roles.include` term occurred in its title or
description, for every source without server-side search. On a Greenhouse
or Lever company board that is noise: every job repeats the same "About
us" text, so a domain term such as "digital health" ("Initrode is a digital
health company…") matched every sales, support and legal job on the board.
`--limit` then cut the concatenated list in source and board order, so the
first noisy boards filled all 50 slots and later boards were never shown.

Aggregators (Remotive, Arbeitnow) differ: each company's boilerplate is one
job among many, and a domain term there is a signal about the employer.

## Decision

- **Greenhouse and Lever** match terms against the title plus the job's
  own labels: Greenhouse `departments[].name`, Lever `categories.team` and
  `categories.department`. The description is never searched there.
- Remotive and Arbeitnow keep title-or-description matching (ADR 0016);
  Adzuna keeps its server-side `what_phrase`.
- **`--limit` is shared round-robin across feeds.** A feed is one
  Greenhouse or Lever board (told apart by company) or one other source.
  Each feed in turn gives its next job until the limit is reached; the kept
  jobs are printed in their original order. With no limit, or fewer jobs
  than the limit, nothing changes.
- No scoring or relevance ordering: ranking stays with the agent.

Rejected: stripping text shared by most of a board's postings (fragile on
small boards and hard to explain); title-only matching everywhere (loses
aggregator recall); ordering title matches first (a form of scoring, and
it still lets one board crowd out the rest).

## Consequences

- Domain words that only appear in boilerplate ("digital health") no
  longer return a whole board; role words ("product manager", "health
  informatics") and department names still do. The skill tells the agent
  to suggest role-like `roles.include` terms when boards return nothing.
- Every board gets a share of the limit, so adding a busy board cannot
  hide the others.
- A role described only in the body of a board job, with a generic title,
  is missed; the user can add that title to `roles.include`.
