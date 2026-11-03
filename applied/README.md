# applied/

Submitted applications. `caddie move <slug> applied` moves the posting
together with its resume, cover letter (Markdown and PDF) and `trace/`
screenshots, sets `applied_at` and `follow_up_after`, and clears
`handoff/`. `caddie follow-ups` reports entries past `follow_up_after`.

## Contributing

Runtime state: everything except this README and `.gitignore` is ignored
and blocked by `scripts/guard-workspace.sh`
([ADR 0024](../docs/adr/0024-workspace-readmes.md)). Fixtures belong in
`testdata/`.
