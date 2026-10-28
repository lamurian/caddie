# 0009. History schema and identity source

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §12.1, §12.2, §12.6, D16, D17

## Context

Verification and tailoring need deterministic parsing of the career history.
Pandoc-style attributes (`{data-icon=…}`) and free-form layouts make parsing
fragile and fork the public CV from the tailored one.

## Decision

- History files are **clean Markdown with a fixed schema**. Each file has one
  `##` section. `###` gives the entry title. The next three non-empty lines
  are the organisation, location and date range
  (`Mon YYYY – Mon YYYY|Present`). Bullets follow until the next `###`.
  `publication.md` and `presentation.md` use the lines after the title for
  venue, authors and date.
- `history/about.md` with YAML frontmatter is the **identity source**: name,
  nationality, work authorisation, location, links and languages, plus the
  long-form `## About` narrative that the agent rewrites per posting.
- Output file names derive from `name`:
  `<First>-<Last>-resume|cover-letter-<position-slug>-<company-slug>.pdf`.

## Consequences

- One source of truth serves both the public CV and the tailored resume.
- Parser errors must name the file and line, so the user can fix their
  Markdown.
- Schema changes touch the parser, `testdata/history/`, `history/README.md`
  and this ADR together.
