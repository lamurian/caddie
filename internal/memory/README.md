# internal/memory/

Caddie's plain-file, user-editable memory in the state directory
(`.caddie/`) ([PRD §12.5](../../PROPOSAL.md),
[ADR 0007](../../docs/adr/0007-plain-file-memory.md),
[ADR 0016](../../docs/adr/0016-job-ids-and-search-memory.md)).

| File | Holds |
|------|-------|
| `seen.go` | `ReadLog`: `seen.jsonl` read leniently; `Latest`, `IDs`, `ByURL`; appending |
| `store.go` | `Store`: `Mark` (`seen add`), `Record` (`posting save`), `Remember`/`Recall` (`last-search.json`) |

`seen.jsonl` is append-only. The last line per ID wins, except that `shown`
never replaces `rejected` or `saved`. A blank, malformed or doubtful line is
skipped and reported as a `Problem` (`file:line: message`), never dropped
silently; a line with an unknown status still counts as seen.

`answers.yml` (phase 04) also lands here.

## Contributing

- Keep the files hand-editable: one JSON object per line, no hidden state.
- Every new tolerance or rejection gets a line in `seen_test.go` asserting
  the reported line number.
- Write only through `caddie` commands; never rewrite a user's lines.
