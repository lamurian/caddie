# 0015. History parser details

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §12.1, §12.2, D17

## Context

ADR 0009 fixes the history schema, but real CVs need a few cases it does not
spell out: entries without a location, degrees dated by year only, wrapped
bullets and author lists. The parser must be deterministic about them and
report errors by file and line.

## Decision

- **Entries:** after `### Title`, the field lines are organisation, an
  optional location, and the date range, which is always the last field
  line. Two field lines mean no location. Fewer than two or more than three
  is an error at the title line.
- **Dates:** `Mon YYYY`, full month names (`August 2022`, `Sept 2014`) and
  `YYYY` are accepted; separators are en dash, em dash or hyphen; the end
  may be `Present` (any case). Output is normalised to `YYYY-MM` or `YYYY`,
  and `present`. A range that ends before it starts is an error.
- **Bullets** start with `- ` or `* `. An indented line after a bullet
  continues it. Any other text after the bullets is an error.
- **Works** (`publication.md`, `presentation.md`): exactly three field lines,
  venue, authors and a single date. Authors are split on `,`, `;` and
  ` and `. Bullets (e.g. a DOI) are allowed.
- Every `*.md` file in `history/` except `about.md` and `README.md` is a
  section, read in file-name order. Its `id` is the file name without `.md`.
- `about.md` must start with frontmatter (strictly decoded, `name`
  required) followed by a non-empty `## About` section. YAML errors keep the
  file's line numbers.
- Parse errors are `profile.ParseError` values (`file:line: message`) that
  match `profile.ErrMalformed`, so the CLI exits 4 and `doctor` fails.

## Consequences

- The parser accepts the common shapes of a hand-written CV without
  attributes or a looser grammar.
- Year-only dates lose month precision; tailoring must not invent months.
- Changing these rules touches the parser, `testdata/history/`,
  `history/README.md` and this ADR together (ADR 0009).
