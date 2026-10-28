# internal/posting/

`caddie posting save <url>`: from a job page to `postings/<slug>.md`
([PRD §6.1, §12.3–12.4](../../PROPOSAL.md),
[ADR 0017](../../docs/adr/0017-posting-extraction.md)).

| File | Holds |
|------|-------|
| `save.go` | `Save`: fetch, merge with the last search, write, mark `saved` in `seen.jsonl` |
| `sections.go` | `Extract`: description blocks → Description, Responsibilities, Requirements, Perks and benefits |
| `files.go` | Finding a saved URL, free slugs (`-2`, `-3`…), `reviews/<company>.md` lookup |

The frontmatter itself is written by `workspace.Posting.Frontmatter`, so
every command writes it the same way. Saving a URL that is already saved
rewrites nothing. The golden file is
`testdata/golden/posting-senior-data-scientist-at-acme.md`
(`go test ./internal/posting -update`).

## Contributing

- A new section cue is one word in `cues` plus a case in `sections_test.go`.
- Never overwrite an existing posting: it may already be tailored.
- Manual-step checklists (`## Manual steps`) arrive with phase 07.
