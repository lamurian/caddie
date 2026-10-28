# 02. Discover: job boards

- **Status:** Todo
- **PRD milestone:** M2 (part 1)

## Goal

From an agent session, "find jobs" works end to end against the public job
board APIs. The agent can search, rank, mark items as shown or rejected, and
save confirmed postings, with any matching reviews merged in.

## Inputs

- PRD §6.1, §7, §12.3–12.5
- ADRs [0001](../adr/0001-skill-plus-cli.md),
  [0007](../adr/0007-plain-file-memory.md),
  [0011](../adr/0011-posting-state-machine.md)

## Scope

- `internal/source`: a `Source` interface and a normalised `Job` struct, with
  adapters for `adzuna`, `remotive`, `arbeitnow`, `greenhouse-board` and
  `lever-board`. HTTP goes behind an interface, and recorded responses live
  in `testdata/sources/`.
- `caddie search [--source s] [--limit n] [--include-seen]` returns
  normalised JSON from the sources enabled in `preferences.yml` and
  de-duplicates by URL.
- `internal/memory`: `seen.jsonl` read/append. `caddie seen add <id…>
  --shown|--rejected` and `caddie seen list [--rejected]`.
- `search` excludes seen IDs by default.
- `internal/posting`: `caddie posting save <url>` fetches the posting,
  extracts its sections (Description, Responsibilities, Requirements, Perks),
  writes `postings/<slug>.md` with frontmatter (`status: saved`), copies
  `reviews/<company>.md` into `## Reviews` and marks the ID as `saved`.
- Skill: the discover workflow (§6.1), with exact commands and the per-item
  ranking fields.

## Out of scope

The `portal` adapter (phase 03). Ranking stays agent judgment and is not
done in code.

## Deliverables

`internal/source/*`, `internal/memory/seen.go`, `internal/posting/*`, CLI
handlers, fixtures and `skills/caddie/references/discover.md`.

## Definition of Done

- [ ] Each adapter has a recorded-response test that maps to `Job`,
      including salary, work mode, location and deadline when present.
- [ ] A disabled source, or missing Adzuna keys, is skipped with a notice,
      not an error.
- [ ] `seen` filtering is tested, both with and without `--include-seen`.
- [ ] `seen.jsonl` tolerates a hand-edited or blank line and reports it.
- [ ] `posting save` produces frontmatter matching PRD §12.3, has a golden
      test, and fills `## Reviews` when a review file exists.
- [ ] Slugs are stable and unique (e.g. `senior-data-scientist-at-acme`).
- [ ] A manual run with real keys returns results (noted in the PR, not run
      in CI).
