# Caddie

A caddie carries the bag, knows the course and advises on the shot, but the
player takes the swing.

Caddie is an AI-assisted job application workflow: a single Go binary plus an
Agent Skill that any coding agent (Pi, Goose, Claude Code…, driven remotely
through Paseo) can load. It:

- discovers postings from EU job boards and international organisations'
  portals, and ranks them against your career history and preferences;
- remembers what you've seen and rejected;
- writes a tailored resume and cover letter in which every claim traces back
  to your `history/`, then renders them to PDF with Quarto;
- fills application forms in a sandboxed headless browser and **stops for
  your approval before submitting**;
- tracks follow-ups and deadlines.

## Status

Pre-alpha. The design is complete and implementation proceeds in the phases
listed in [`docs/tasks/`](docs/tasks/README.md).

## Documentation

- [`PROPOSAL.md`](PROPOSAL.md): the full product requirements document.
- [`AGENTS.md`](AGENTS.md): contributor and agent entry point, rules and
  directory map.
- [`docs/adr/`](docs/adr/README.md): architecture decisions.
- [`docs/tasks/`](docs/tasks/README.md): development phases.

## Quick setup (once phase 01 lands)

```bash
cp .env.example .env
cp preferences.example.yml preferences.yml
cp answers.example.yml answers.yml
go build ./cmd/caddie && ./caddie doctor
```

## Licence

MIT. See [LICENSE](LICENSE).
