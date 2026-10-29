# internal/document/

Checks the agent's tailored Markdown drafts (`resumes/<slug>.md`,
`cover-letters/<slug>.md`) before they are rendered
([PRD §6.2–6.3](../../PROPOSAL.md),
[ADR 0004](../../docs/adr/0004-fidelity-check.md)).

| File | Holds |
|------|-------|

## Contributing

- Matching stays deterministic and conservative: when in doubt, report a
  claim as untraceable rather than let an invented one through.
- Every new claim kind or tolerance gets a table-driven case in the tests,
  with a fixture in [`testdata/documents/`](../../testdata/documents/README.md).
- Golden reports are regenerated with `go test ./internal/document -update`.
