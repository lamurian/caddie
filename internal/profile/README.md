# internal/profile/

Parses `history/` into a `Profile` that serialises to JSON
([ADR 0009](../../docs/adr/0009-history-schema.md),
[ADR 0015](../../docs/adr/0015-history-parser-details.md)).

| File | Holds |
|------|-------|
| `load.go` | `Load(dir)`: reads every `*.md` except `README.md`, skips subdirectories |
| `about.go` | `about.md` YAML frontmatter → `Identity` (name, links, languages, …) |
| `section.go` | One history file → `Section` of `Entry` (title, organisation, location, dates, bullets); `SplitAuthors` |
| `dates.go` | `ParseDate`, `ParseRange`: month names, `Sept`, year-only, `Present`, `-` / `–` / `—` |
| `types.go` | `Profile`, `Section`, `Entry`, `Work` (publications, presentations) |

Every problem is a `*ParseError` that prints `file:line: message` and wraps
`ErrMalformed`, which the CLI maps to exit code 4.

## Contributing

- A schema change starts in `history/README.md` and an ADR, then the parser,
  then `testdata/history/` and the golden file
  (`go test ./cmd/caddie -update`).
- Add a table-driven error case for every new rejection, asserting the line.
- Contact details (phone, email) never live here; they are in
  `preferences.yml`.
