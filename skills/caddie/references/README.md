# skills/caddie/references/

Workflow references that `SKILL.md` links to and the agent loads on demand,
one file per workflow ([ADR 0012](../../../docs/adr/0012-agentic-docs-and-line-limits.md)).

| File | Workflow |
|------|----------|
| [`discover.md`](discover.md) | Find, rank and save jobs (PRD §6.1, phases 02–03) |
| [`tailor.md`](tailor.md) | Draft, verify and keyword-check a resume (PRD §6.2, phase 04) |
| [`cover-letter.md`](cover-letter.md) | Answers, interview, draft and verify a cover letter (PRD §6.3, phase 04) |
| [`browse.md`](browse.md) | Drive the sandboxed browser: sessions, snapshots, uploads, logins (PRD §7, §9, phase 06) |
| [`apply.md`](apply.md) | Fill a form, the submit gate, submit or manual route, `caddie move` (PRD §6.4–6.5, phase 07) |
| [`exit-codes.md`](exit-codes.md) | The exit-code table (mirrored from `cmd/README.md`) and the codes each command returns (phase 07) |

## Contributing

- Written for the agent: numbered steps, exact commands, no rationale.
- Every command shown must exist in the CLI with exactly that syntax.
- Under 100 lines each; link every new file from `SKILL.md` and this table.
