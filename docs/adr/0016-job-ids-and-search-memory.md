# 0016. Job IDs, search filtering and search memory

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §6.1, §7, §12.5, D13

## Context

`caddie search`, `caddie seen` and `caddie posting save` must agree on what
identifies a job, and `seen.jsonl` lines need a URL, title and company even
though the agent passes only IDs to `seen add`. The PRD leaves open how
boards without a search parameter are narrowed, what happens when one board
is down, and how statuses combine when a job is marked twice.

## Decision

- **Job IDs** are `<source>:<native id>`: `adzuna:<id>`, `remotive:<id>`,
  `arbeitnow:<slug>`, `greenhouse-board:<job id>`, `lever-board:<uuid>`.
  A page no adapter knows gets `url:` plus the first 12 hex digits of the
  SHA-256 of its URL. `posting save` reuses the ID the last search or
  `seen.jsonl` recorded for the URL, else derives one from the URL.
- **Search terms** are `roles.include`. Adzuna matches them server side
  (`what_phrase`, one request per country and term). Remotive is queried
  once per term (`search`), but its matching is loose, so its results, like
  Arbeitnow's, Greenhouse's and Lever's, are kept only when a term occurs
  in the title or description. `roles.exclude` (title substring) and
  `companies.exclude` (exact name) are applied to every source. All
  matching is case-insensitive. Ranking stays with the agent.
- Results are **de-duplicated by URL**, first source wins, in the fixed
  order adzuna, remotive, arbeitnow, greenhouse-board, lever-board. Seen IDs
  are hidden and counted in `hidden_seen` unless `--include-seen`.
  `--limit` defaults to 50; `summary` is plain text capped at 600 runes.
- A disabled or unconfigured source, a failing source and a malformed
  `seen.jsonl` line are **notices** in the JSON, not errors. Only when every
  queried source fails does `search` exit 8. Error text never contains a
  query string, because Adzuna keys travel in it.
- **Search memory:** `search` writes the jobs it printed to
  `.caddie/last-search.json` (newest first, merged with earlier searches,
  at most 1000 jobs). `seen add` and `posting save` read details from it;
  an ID it does not know is still recorded and listed as `unknown`.
- **Status precedence:** `seen.jsonl` is append-only and the last line per
  ID wins, but `shown` never replaces `rejected` or `saved`, so showing a
  job again cannot undo a decision. Repeating a status appends nothing.
- `--json` is accepted and always on for `search`, `seen` and
  `posting save`, like `profile` (ADR 0014). Exit 5 now also means "no job
  posting at that URL"; exit 8 means an HTTP failure.

## Consequences

- IDs are stable across searches, readable in `seen.jsonl` and safe to
  pass on a command line.
- Arbeitnow and the boards download whole listings to filter them locally;
  that is cheap for one page or one company board.
- `last-search.json` is a cache: deleting it loses only the details of
  future `seen add` lines, never a decision.
- Changing the ID scheme orphans existing `seen.jsonl` lines, so it needs a
  new ADR.
