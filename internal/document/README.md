# internal/document/

Checks the agent's tailored Markdown drafts (`resumes/<slug>.md`,
`cover-letters/<slug>.md`) before they are rendered
([PRD §6.2–6.3](../../PROPOSAL.md),
[ADR 0004](../../docs/adr/0004-fidelity-check.md)).

| File | Holds |
|------|-------|
| `keywords.go` | `Keywords(posting, draft, profile)` → `KeywordReport`, its statuses and `Text()` |
| `keywords_extract.go` | `## Requirements` → de-duplicated keyword phrases, filler stripping |
| `keywords_match.go` | Folding, tokens, plural stem, phrase boundaries, the history text |
| `golden_test.go` | The `-update` flag and `checkGolden` shared by every golden test here |

## Keywords

`Keywords` reads the lines under the posting's `## Requirements` (frontmatter,
labels like `**What you bring**` or `Nice to have:` and `###` sub-headings
skipped), drops a `Label:` prefix and splits on `,` `;` `/` `|` `( )` `. `
and ` and ` / ` or ` / ` & `. Each phrase is lowercased and stripped, until
stable, of `N+ years of`, lead-ins up to `skills|experience|knowledge|… in|with|of|using`,
filler prefixes (`strong`, `fluent`, `ability to`, …) and suffixes (`skills`,
`is a plus`, `preferred`, …). Phrases of only noise words (`related field`)
are dropped; phrases over 5 words are skipped with a notice; duplicates (same
match tokens) keep the first. Matching is on folded tokens: lowercase,
diacritics removed, hyphens as spaces, `a/an/the/of/in/on/for/to/with/at/by`
ignored, a final `s` dropped (not `-ss/-us/-is`), whole words only and never
across punctuation, a blank line or a new bullet. History is `about`,
language names, section headings, entry titles, organisations and bullets,
and work titles, venues and bullets. Status: `covered` (draft and history),
`missing` (history only: may be added), `unsupported` (neither: never add),
`unbacked` (draft only: run `caddie verify` and remove it).

## Contributing

- Matching stays deterministic and conservative: when in doubt, report a
  claim as untraceable rather than let an invented one through.
- Every new claim kind or tolerance gets a table-driven case in the tests,
  with a fixture in [`testdata/documents/`](../../testdata/documents/README.md).
- Golden reports are regenerated with `go test ./internal/document -update`.
