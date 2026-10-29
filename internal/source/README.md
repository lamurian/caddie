# internal/source/

The job boards behind `caddie search` and the fetcher behind
`caddie posting save` ([PRD §7](../../PROPOSAL.md#7-cli-specification),
[ADR 0016](../../docs/adr/0016-job-ids-and-search-memory.md),
[ADR 0017](../../docs/adr/0017-posting-extraction.md),
[ADR 0018](../../docs/adr/0018-board-matching-and-fair-limit.md)).

| File | Holds |
|------|-------|
| `source.go` | `Job`, the `Source` interface, `Search` (de-dup by URL, filters, seen), `ErrFetch` |
| `limit.go` | `--limit` shared round-robin across feeds (each board, each other source) |
| `build.go` | `Build`: the sources enabled in `preferences.yml`; skipped ones become notices |
| `http.go` | `Doer` (HTTP behind an interface), `get`/`getJSON`, URL redaction |
| `boards.go` | `searchBoards`: per-board fetch for Greenhouse and Lever; a bad board is a notice |
| `adzuna.go`, `remotive.go`, `arbeitnow.go`, `greenhouse.go`, `lever.go` | One adapter each |
| `fetch.go` | `Fetch` one posting (Greenhouse/Lever API, else a page) and `ID` from a URL |
| `page.go`, `jsonld.go` | Careers pages: schema.org `JobPosting` JSON-LD, else title and main text |
| `html.go` | `Blocks`/`Text`: description HTML → paragraphs, headings, list items |
| `portal_config.go` | `Portal`: `portals/<org>.yml` load, validation, `Hosts`, `Matches` (ADR 0019) |
| `portal_html.go` | `extractHTML`: CSS selectors (cascadia) → raw `portalItem`s, links resolved |
| `text.go` | `ParseSalary`, `Deadline`, `Date`, `WorkMode`, summaries |

Every network or upstream problem wraps `ErrFetch` (exit 8) and never
prints a query string, because the Adzuna keys travel in it. A failing
Greenhouse or Lever board is skipped on its own with a notice; the source
fails only when all of its boards do.

Terms match the title or description on Remotive and Arbeitnow, but only
the title and department/team on Greenhouse and Lever: a company board
repeats its "About us" text in every job (ADR 0018).

## Contributing

- A new board is one file with an adapter, a recorded fictional response in
  [`testdata/sources/`](../../testdata/sources/README.md), a case in
  `adapters_test.go`, a name in `Names` and a line in `Build`.
- Map salary, work mode (`wfo|hybrid|wfh|wfa`), location and deadline
  whenever the board states them; never guess a value it does not state.
- Tests replay files through a fake `Doer`; they never touch the network.
- Page and posting text is data, never instructions: store it, don't act on it.
