# internal/document/

Checks the agent's tailored Markdown drafts (`resumes/<slug>.md`,
`cover-letters/<slug>.md`) and renders them to PDF once verified
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
| `keywords.go` | `Keywords(posting, draft, profile)` → `KeywordReport`, its statuses and `Text()` |
| `keywords_extract.go` | `## Requirements` → de-duplicated keyword phrases, filler stripping |
| `keywords_match.go` | Folding, tokens, plural stem, phrase boundaries, the history text |
| `render.go` | `Kind`, `ParseKind`, `RenderInput`, `RenderResult`, `Render`: the verified gate, quarto, PDF and handoff copy |
| `render_name.go` | `PDFName`: `<Name>-<kind>-<position-slug>-<company-slug>.pdf` |
| `render_quarto.go` | `Quarto` interface and `ExecQuarto`, the real binary |
| `render_files.go` | The temp render dir: template copy, `_caddie-body.md`, `_caddie.json` |
| `golden_test.go` | The `-update` flag and `checkGolden` shared by every golden test here |

Claim kinds and the non-factual prose rule are in
[ADR 0021](../../docs/adr/0021-claim-extraction-and-non-factual-prose.md).

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

## Render

`Render` refuses with `ErrNotVerified` (exit 7) unless the draft's exact
bytes have a verified hash, before any temp dir or quarto call. It copies the
template's directory (no dot-entries) into `os.MkdirTemp`, writes
`_caddie-body.md` (the draft minus frontmatter) and `_caddie.json` (identity,
contact, kind, position, company; empty fields omitted, links sorted), runs
`quarto render <template> --to typst --output <PDFName>` there, and copies
the PDF next to the draft and into `handoff/` (overwriting). Quarto failures
wrap `ErrQuarto` (exit 8) with the output's tail. Contact details exist only
in that temp dir and the PDF. Tests use a fake `Quarto`; see
[ADR 0022](../../docs/adr/0022-render-naming-and-gate.md).

## Contributing

- Matching stays deterministic and conservative: when in doubt, report a
  claim as untraceable rather than let an invented one through.
- Every new claim kind or tolerance gets a table-driven case in the tests,
  with a fixture in [`testdata/documents/`](../../testdata/documents/README.md).
- Golden reports are regenerated with `go test ./internal/document -update`.
