# internal/memory/

Caddie's plain-file, user-editable memory: the state directory
(`.caddie/`) and `answers.yml` ([PRD §12.5](../../PROPOSAL.md),
[ADR 0007](../../docs/adr/0007-plain-file-memory.md),
[ADR 0016](../../docs/adr/0016-job-ids-and-search-memory.md),
[ADR 0020](../../docs/adr/0020-answers-scopes-and-form-keys.md)).

| File | Holds |
|------|-------|
| `seen.go` | `ReadLog`: `seen.jsonl` read leniently; `Latest`, `IDs`, `ByURL`; appending |
| `store.go` | `Store`: `Mark` (`seen add`), `Record` (`posting save`), `Remember`/`Recall` (`last-search.json`) |
| `answers.go` | `ParseScope`, `LoadAnswers`/`ParseAnswers`, `For` (merged view), `All`, `FormAnswers` |
| `answers_edit.go` | Shape validation (`ErrMalformedAnswers`), `Set` on the YAML node tree, atomic `Save` |

`seen.jsonl` is append-only. The last line per ID wins, except that `shown`
never replaces `rejected` or `saved`. A blank, malformed or doubtful line is
skipped and reported as a `Problem` (`file:line: message`), never dropped
silently; a line with an unknown status still counts as seen.

`answers.yml` is a YAML node tree: `Set` edits it in place so key order
and hand comments survive `Save`. Values may be any scalar (kept as
literal text); a structural problem is an `ErrMalformedAnswers` naming
`file:line`. Lookup is posting → company → global; an empty value is
unanswered. `memory` never imports `workspace`: callers pass slugs.

## Contributing

- Keep the files hand-editable: one JSON object per line, no hidden state.
- Every new tolerance or rejection gets a line in `seen_test.go` or
  `answers_test.go` asserting the reported line number.
- Write only through `caddie` commands; never rewrite a user's lines.
