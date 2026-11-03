# internal/workspace/

The workspace's shape and the posting lifecycle
([ADR 0008](../../docs/adr/0008-workspace-tracking.md),
[ADR 0011](../../docs/adr/0011-posting-state-machine.md)).

| File | Holds |
|------|-------|
| `layout.go` | `Layout`: paths of `postings/`, `applied/`, `manual/`, … and `Missing()` |
| `slug.go` | `Slug` / `PostingSlug`: ASCII, lower-case, accents stripped (`<title>-at-<company>`) |
| `posting.go` | `Posting`, the YAML frontmatter of `postings/<slug>.md`, and `ParsePosting` |
| `frontmatter.go` | `Posting.Frontmatter`: writes it back in PRD §12.3 order, unset values empty |
| `state.go` | `Status` and `Transition(from, to)`: the only legal status changes |
| `advance.go` | `StatusPath` (the single steps from → to) and `AdvanceStatus`: rewrites a posting's `status` atomically, body byte for byte; same status → file untouched |
| `move.go` | `Move`, `MoveOptions`, `MoveResult`, `ErrDestExists`, `ErrBadMove`: preconditions and frontmatter of `caddie move` ([ADR 0026](../../docs/adr/0026-move-archive-and-trace.md)) |
| `move_plan.go` | Which artefacts exist and land in `<dest>/<slug>/`; `ClearHandoff` |
| `move_exec.go` | The journal: renames (copy + remove across devices), posting written last, rollback in reverse |

`AdvanceStatus` changes only the frontmatter in place. `Move` relocates the
posting to `<dest>/<slug>.md` and its artefacts to `<dest>/<slug>/`; every
check runs first and a failure mid-move is undone. Afterwards it deletes
the files in `handoff/` (not the tracked `.gitignore` and `README.md`, not subdirectories) and leaves
the trace directory present but empty.

## Contributing

- Add a status or transition only with an ADR amending 0011, and extend
  `state_test.go`, which checks every legal pair and rejects all others.
- Frontmatter fields follow PRD §12.3; add them to the fixture in
  `testdata/postings/` with the code.
- Move tests swap the package vars `rename`, `writePosting` and
  `removeFile` to inject failures; they must not run in parallel.
- Illegal transitions wrap a sentinel the CLI maps to exit code 6.
