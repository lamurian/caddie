# skills/caddie/

The Agent Skill itself. An agent loads [`SKILL.md`](SKILL.md): what Caddie
is, the directory layout, the hard rules (PRD §8.3, verbatim), the gate
message and the exit-code table. This README is for contributors; agents
read `SKILL.md`.

`references/` (one file per workflow) is added when the first workflow
lands; see the layout in [`skills/README.md`](../README.md).

## Contributing

- `SKILL.md` stays under 100 lines; move detail into `references/`.
- The exit-code table is mirrored verbatim from `cmd/README.md`; a test in
  `cmd/caddie` fails if they differ.
- Every command shown must exist in the CLI with exactly that syntax.
