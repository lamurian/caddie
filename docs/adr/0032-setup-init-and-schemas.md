# 0032. First-run setup: `caddie init`, `validate` and JSON schemas

- **Status:** Accepted
- **Date:** 2026-11-03
- **PRD refs:** §12.5, §12.7, §13

## Context

Setup was three `cp *.example.yml` steps and hand edits. Nothing checked
`answers.yml` up front, and a fresh user had no guide to which keys matter.
Two agents asked to "set up preferences" would ask different questions
and write different files: that is mechanics, so it belongs in the binary
([ADR 0001](0001-skill-plus-cli.md)).

## Decision

- **One question set, in code.** `internal/setup` holds the default
  questions (id = dotted YAML path, file, prompt, kind, default, options).
  `apply.require_gate` is never asked; it is always written `true`.
- **`caddie init`** copies `.env.example` to `.env` if missing. Then, for
  each of `preferences.yml` and `answers.yml` that is missing, it asks the
  questions on stdin, writes the file (mode 0600, never overwriting) and
  validates it. Existing files are validated and left alone.
  - `--questions` prints the questions for the missing files as JSON and
    changes nothing; the agent asks them (plus its own) in chat.
  - `--answers <file|->` takes the replies as JSON keyed by question id;
    an unknown id or a reply that does not fit is exit 2.
- **`caddie validate [preferences|answers]`** prints one `ok`/`fail` line
  per file. A missing file, invalid preferences or `require_gate: false`
  is exit 3; a malformed `answers.yml` is exit 4. `caddie doctor` gains an
  `answers` line: `warn` while missing, `fail` when malformed.
- **Schemas.** `schemas/preferences.schema.json` and
  `schemas/answers.schema.json` (JSON Schema 2020-12) document both files
  for editors via a `# yaml-language-server: $schema=…` header. The Go
  loaders stay the validator, so no schema library is added; tests keep
  the schema's keys and enums equal to the Go structs and validators.
  Preferences also now reject negative salaries and day counts.
- **mise.** `mise run compile` installs the binary (`go install`);
  `mise run init` compiles, runs `caddie init` and `caddie validate`, then
  an advisory `caddie doctor`.
- **Skill.** `references/setup.md`: on first run, `caddie validate`; if a
  file is missing, ask the default questions and write via
  `caddie init --answers`.

## Consequences

- The terminal and the agent ask the same questions and write the same
  files; adding a question is a code change with a test.
- The schema files are hand-written; a key added to `Preferences` fails
  the sync test until the schema is updated.
- `caddie init` never rewrites a file; to start again, delete it first.
