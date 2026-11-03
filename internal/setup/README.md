# internal/setup/

First-run setup behind `caddie init` and `caddie validate`
([ADR 0032](../../docs/adr/0032-setup-init-and-schemas.md)).

- **`questions.go`**: the default question set, shared by the terminal Q&A
  and the agent (`caddie init --questions`). A question's `ID` is the
  dotted YAML path it fills; `File` is `preferences` or `answers`.
  `apply.require_gate` is never asked: it is always `true`.
- **`values.go`**: `Parse` reads typed text (empty means the default;
  lists are comma-separated; choices must be options); `FromJSON` reads
  the agent's `--answers` JSON and rejects unknown ids. Errors wrap
  `ErrAnswer` (exit 2).
- **`build.go`**: `Preferences` and `Answers` render the files, with the
  `$schema` header, and validate them with `config.ParsePreferences` and
  `memory.ParseAnswers`.
- **`prompt.go`**: `Ask`, the line-based prompt loop; it re-asks until a
  reply fits.
- **`check.go`**: `CheckPreferences` (also refusing `require_gate: false`),
  `CheckAnswers`, and `Write`, which never overwrites a file.

## Contributing

- A new question needs its key in `schemas/`, the Go struct (for
  preferences) and a case in `setup_test.go`.
- Never ask for secrets here: API keys go in `.env`, by hand.
- Keep the set short; the agent adds role-specific questions itself.
