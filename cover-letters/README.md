# cover-letters/

Tailored cover-letter drafts (`<slug>.md`) and the PDFs rendered from them.
A draft renders only after `caddie verify` passes and stays unchanged
([ADR 0004](../docs/adr/0004-fidelity-check.md),
[0022](../docs/adr/0022-render-naming-and-gate.md)). The PDF is also copied
to `handoff/` for upload.

## Contributing

Runtime state: everything except this README and `.gitignore` is ignored
and blocked by `scripts/guard-workspace.sh`
([ADR 0024](../docs/adr/0024-workspace-readmes.md)). Fixtures belong in
`testdata/`.
