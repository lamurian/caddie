# internal/document/

Checks the agent's tailored Markdown drafts (`resumes/<slug>.md`,
`cover-letters/<slug>.md`) before they are rendered
([PRD §6.2–6.3](../../PROPOSAL.md),
[ADR 0004](../../docs/adr/0004-fidelity-check.md)).

| File | Holds |
|------|-------|
| `verify.go` | `Verify`, `Sources`, `Claim`, `Report`: walks frontmatter, headings, skills section, entries and prose |
| `verify_entry.go` | A `###` block against one history entry or work (title, fields, date range) |
| `verify_prose.go` | Prose lines (URLs, dates, numbers, term phrases) and skills items |
| `verify_corpus.go` | What a draft may claim: profile, answers, targets |
| `verify_dates.go` | Dates in free text → `YYYY-MM-DD`, `YYYY-MM`, `YYYY` via `profile.ParseDate` |
| `verify_words.go` | Word splitting, case and diacritics folding, numbers, term shape, allowlist |
| `hash.go` | `Hash`, `RecordVerified`, `IsVerified`: `.caddie/verified/<sha256>` |

Claim kinds and the non-factual prose rule are in
[ADR 0021](../../docs/adr/0021-claim-extraction-and-non-factual-prose.md).

## Contributing

- Matching stays deterministic and conservative: when in doubt, report a
  claim as untraceable rather than let an invented one through.
- Every new claim kind or tolerance gets a table-driven case in the tests,
  with a fixture in [`testdata/documents/`](../../testdata/documents/README.md).
- Golden reports are regenerated with `go test ./internal/document -update`.
