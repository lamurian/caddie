# internal/

All Go logic. Packages are deterministic, make no LLM calls and are unit
tested against `testdata/` ([ADR 0001](../docs/adr/0001-skill-plus-cli.md)).

## Packages

| Package | Responsibility | PRD |
|---------|----------------|-----|
| [`config`](config/README.md) | Load `.env`, env vars (env wins), `preferences.yml` | §12.7–12.8 |
| [`workspace`](workspace/README.md) | Directory layout, slugs, posting state machine, `Move` | §6.5, §12.3 |
| [`profile`](profile/README.md) | Parse `history/*.md` + `about.md` into a struct → JSON | §12.1–12.2 |
| [`doctor`](doctor/README.md) | Health checks behind `caddie doctor`; reports, never changes | §7 |
| [`posting`](posting/README.md) | Fetch, extract, frontmatter, reviews merge, manual checklist | §6.1, §12.3–12.4 |
| [`source`](source/README.md) | Adapters: adzuna, remotive, arbeitnow, greenhouse, lever, portal | §7 |
| [`memory`](memory/README.md) | `.caddie/seen.jsonl`, the last-search cache, `answers.yml` | §12.5 |
| [`document`](document/README.md) | `Verify`, `Keywords`, `Render` (Quarto) + verified hashes | §6.2–6.3 |
| `browse` | agent-browser argv builder behind an `Exec` interface | §9 |

## Dependency direction

`cmd` → any package. `workspace`, `posting`, `document` may use `config`,
`profile` and `memory`; `posting` also uses `source` and `workspace`;
`document` also uses `workspace` (for `Slug` in PDF names).
`doctor` may use every package it checks. `source` uses only `config`,
returns plain structs and never touches the filesystem layout; `memory`
stores `source.Job` values in its search cache. `browse` depends only on `config`. No cycles; no package
imports `cmd`.

## Conventions

- One package per directory, with a `README.md` (linked in the table above)
  and a `doc.go` package comment. Files
  under 300 lines (tests included). Split by concern, e.g.
  `source/adzuna.go`, `source/lever.go`.
- External processes (`quarto`, `docker compose`, `agent-browser`) and HTTP
  go behind small interfaces so that tests can use fakes.
- Errors carry enough context for the CLI to pick an exit code; define
  sentinel errors per package.
- Tests are table-driven, use `testdata/` fixtures and golden files, and
  never touch the network.

## Contributing

1. Check which phase in [`docs/tasks/`](../docs/tasks/README.md) owns the
   package and stay inside its scope.
2. Write the test first against a fixture in `testdata/`.
3. Keep exported APIs small; `cmd/` should need one call per subcommand.
4. If you add a package, give it a `README.md` and add it to the table above.
