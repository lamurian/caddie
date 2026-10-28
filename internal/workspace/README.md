# internal/workspace/

The workspace's shape and the posting lifecycle
([ADR 0008](../../docs/adr/0008-workspace-tracking.md),
[ADR 0011](../../docs/adr/0011-posting-state-machine.md)).

| File | Holds |
|------|-------|
| `layout.go` | `Layout`: paths of `postings/`, `applied/`, `manual/`, … and `Missing()` |
| `slug.go` | `Slug` / `PostingSlug`: ASCII, lower-case, accents stripped (`<title>-at-<company>`) |
| `posting.go` | `Posting`, the YAML frontmatter of `postings/<slug>.md`, and `ParsePosting` |
| `state.go` | `Status` and `Transition(from, to)`: the only legal status changes |

Phase 01 only validates transitions; moving files between directories
(`caddie move`) arrives in a later phase.

## Contributing

- Add a status or transition only with an ADR amending 0011, and extend
  `state_test.go`, which checks every legal pair and rejects all others.
- Frontmatter fields follow PRD §12.3; add them to the fixture in
  `testdata/postings/` with the code.
- Illegal transitions wrap a sentinel the CLI maps to exit code 6.
