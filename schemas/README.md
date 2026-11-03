# schemas/

JSON Schemas (draft 2020-12) for the two private config files
([ADR 0032](../docs/adr/0032-setup-init-and-schemas.md)):

| File | Describes |
|------|-----------|
| [`preferences.schema.json`](preferences.schema.json) | `preferences.yml`: search, ranking and `apply` rules (PRD §12.7) |
| [`answers.schema.json`](answers.schema.json) | `answers.yml`: `global`, `company`, `posting` answers ([ADR 0007](../docs/adr/0007-plain-file-memory.md)) |

Editors with the YAML language server pick them up from the
`# yaml-language-server: $schema=…` first line that `caddie init` and the
example files carry. `caddie validate` enforces the same rules with the Go
loaders (`internal/config`, `internal/memory`), so no schema library is
needed at runtime.

## Contributing

- Change a schema together with the Go struct or validator it mirrors.
  `internal/config/schema_test.go` and `internal/memory/schema_test.go`
  fail when keys or enums drift apart.
- Keep `additionalProperties: false` on every preferences object: unknown
  keys are errors in the loader too.
- `apply.require_gate` stays `const: true` ([ADR 0003](../docs/adr/0003-human-submit-gate.md)).
