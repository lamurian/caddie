# history/

The user's career history. It is **tracked and public**: it doubles as a public
CV and is the single source of truth for every resume and cover letter
([ADR 0009](../docs/adr/0009-history-schema.md)). `caddie verify` rejects any
claim that cannot be traced to these files or to a stored answer.

## Files

`about.md` (identity + narrative), `employment.md`, `consultancies.md`,
`education.md`, `research.md`, `teaching.md`, `publication.md`,
`presentation.md`.

## Schema (clean Markdown, no attributes like `{data-icon=…}`)

Each file has one `##` heading, followed by entries:

```markdown
## Employment

### External Consultant
INA-RESPOND
Jakarta, Indonesia
Aug 2022 – Jun 2023
- Streamlined data processing and analysis pipelines
```

- `###` is the title. The next three non-empty lines are the organisation,
  the location and the date range (`Mon YYYY – Mon YYYY|Present`). Bullets
  follow until the next `###`.
- In `publication.md` and `presentation.md`, the lines after `###` are the
  venue, authors (split on `,`, `;`, ` and `) and date.
- The location line may be left out. Dates may also be `YYYY` or a full
  month name; an indented line continues the bullet above it
  ([ADR 0015](../docs/adr/0015-history-parser-details.md)).
- Check your files with `caddie doctor`: errors name the file and line.

`about.md` has YAML frontmatter:

| Field | Values |
|-------|--------|
| `name` | Used to name the generated PDF files |
| `nationality` | Country code |
| `work_authorization` | List of country or region codes |
| `location` | City, country |
| `links` | github, scholar, orcid |
| `languages` | `{name, level}` pairs |

The body is a `## About` narrative.

## Never put here

Phone, email, salary or anything else not meant for publication. Contact
details live in `preferences.yml` (ignored) and are merged in at render time.

## Contributing

Use the user's own data only. Test data belongs in `testdata/history/`.
Changing the schema means updating the `profile` parser, its fixtures and
this README in the same change.
