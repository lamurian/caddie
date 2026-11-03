# testdata/golden/

Expected outputs that tests compare byte for byte.

| File | Produced by | Regenerate |
|------|-------------|------------|
| `profile.json` | `caddie profile --json` on `testdata/` | `go test ./cmd/caddie -update` |
| `posting-senior-data-scientist-at-acme.md` | `posting.Save` on `sources/greenhouse-acme-123.json` + `reviews/acme.md`; `cmd/caddie` checks it too | `go test ./internal/posting -update` |
| `render-resume.txt` | `pdftotext -layout` of `document.Render` (real quarto) on `documents/resume-traceable.md`, `history/` and `config/preferences.fixture.yml`; trailing blanks and blank lines dropped | `go test ./internal/document -run Real -update` (needs `quarto`, `pdftotext`) |
| `render-cover-letter.txt` | The same for `documents/cover-letter-answers.md` | as above |
| `keywords-lead-biostatistician-at-globex.txt` | `document.Keywords(...).Text()` on `postings/lead-biostatistician-at-globex.md`, `documents/keywords-draft-lead-biostatistician.md` and `history/` | `go test ./internal/document -update` |

## Contributing

Never hand-edit a golden file to make a test pass. Change the code or the
fixture, regenerate with `-update`, and review the diff like code.
