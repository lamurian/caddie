# reviews/

Notes **the user** keeps about employers, one file per company:
`reviews/<company-slug>.md`. Caddie never writes here and never scrapes
review sites; it only reads what you put in.

## Format (PRD §12.4)

```markdown
---
company: Acme
updated: 2026-10-20
---
- Glassdoor (2026-09): 3.8/5, praised flexibility, criticised promotion pace
- Colleague note: team in Amsterdam is small, remote-friendly in practice
```

## How it is used

`caddie posting save <url>` copies the matching file into the
new posting's `## Reviews` section, so the summary is shown during ranking and
stays with the application. Edit the file any time; later saves pick it up.

## Contributing

Personal data: everything except this README and `.gitignore` is ignored
([ADR 0024](../docs/adr/0024-workspace-readmes.md)). Fixtures belong in
`testdata/`.
