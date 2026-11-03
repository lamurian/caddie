# postings/

Postings under consideration, one file per job: `<slug>.md` with YAML
frontmatter (PRD §12.3). `caddie posting save <url>` creates them with
`status: saved`; tailoring and rendering advance the status
([ADR 0011](../docs/adr/0011-posting-state-machine.md)).
`caddie move` takes them to `applied/` or `manual/`.

## Contributing

Runtime state: everything except this README and `.gitignore` is ignored
and blocked by `scripts/guard-workspace.sh`
([ADR 0024](../docs/adr/0024-workspace-readmes.md)). Fixtures belong in
`testdata/`.
