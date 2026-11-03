# internal/posting/

`caddie posting save <url>`: from a job page to `postings/<slug>.md`
([PRD §6.1, §12.3–12.4](../../PROPOSAL.md),
[ADR 0017](../../docs/adr/0017-posting-extraction.md)).

| File | Holds |
|------|-------|
| `save.go` | `Save`: fetch, merge with the last search, write, mark `saved` in `seen.jsonl` |
| `sections.go` | `Extract`: description blocks → Description, Responsibilities, Requirements, Perks and benefits |
| `files.go` | Finding a saved URL, free slugs (`-2`, `-3`…), `reviews/<company>.md` lookup |
| `manual.go` | `SetManualSteps` / `MergeManualSteps`: the `## Manual steps` checklist |

The frontmatter itself is written by `workspace.Posting.Frontmatter`, so
every command writes it the same way. Saving a URL that is already saved
rewrites nothing. The golden file is
`testdata/golden/posting-senior-data-scientist-at-acme.md`
(`go test ./internal/posting -update`).

## Portal URLs

`Options.Portals` holds every `portals/*.yml`, loaded by the CLI whether or
not `preferences.yml` enables it ([ADR 0019](../../docs/adr/0019-portal-config-schema.md)).
A URL a config matches (its `detail_url`, else its listing host) is saved
with source `portal:<org>`, ID `portal:<org>:<id>` (the URL hash when the
pattern captures no id), `ats` from the config and the organisation's
`name` as company unless the page names another. The page itself is read
like any careers page. An invalid config fails every save with exit 3,
naming the file: config typos are loud (ADR 0014). Without a remembered
search, `seen.jsonl` restores the source from the ID (`source.SourceOf`).

## Manual steps

`SetManualSteps(path, steps)` (behind `caddie posting steps`,
[ADR 0024](../../docs/adr/0024-move-archive-and-trace.md)) works on a
posting in `postings/` or `manual/`. Frontmatter bytes are kept as written;
only the body changes. A missing section is appended at the end; an
existing one keeps its items (order, text, tick) and other lines, and new
`- [ ]` items go after its last item. Steps match trimmed,
whitespace-collapsed and case-insensitive, so a ticked item is never
re-added or unticked and a second run rewrites nothing. `ErrNoSteps` and
`ErrEmptyStep` are usage errors.

## Contributing

- A new section cue is one word in `cues` plus a case in `sections_test.go`.
- Never overwrite an existing posting: it may already be tailored.
