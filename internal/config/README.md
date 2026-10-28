# internal/config/

Loads Caddie's settings ([PRD §12.7–12.8](../../PROPOSAL.md)).

- **`.env` and env vars** (`config.go`, `dotenv.go`): `Load` reads `.env`
  from the current directory, then real environment variables. Precedence is
  env var > `.env` > built-in defaults; an empty value falls back to the
  default. Defaults equal `.env.example` (a test checks this). A malformed
  line is an error naming the file and line; an unknown key is an error too.
- **`preferences.yml`** (`preferences.go`): `LoadPreferences` parses it
  strictly (typos are errors). `apply.require_gate` defaults to `true` when
  omitted. The `contact` block never serialises to JSON, so phone and email
  cannot leak by accident.
- `Config.Resolve` turns a path relative to `CADDIE_WORKSPACE` into a real path.

Every error wraps `ErrConfig` or `ErrNoPreferences`, which the CLI maps to
exit code 3 ([ADR 0014](../../docs/adr/0014-cli-contract.md)).

## Contributing

- A new setting goes in `.env.example` (or `preferences.example.yml`), the
  defaults here and a test, in the same change.
- Never log or print secret values; never add a real value to a fixture.
- Never let a default weaken the submit gate
  ([ADR 0003](../../docs/adr/0003-human-submit-gate.md)).
