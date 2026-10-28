# 0012. Agentic docs and line limits

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** — (project convention)

## Context

Coding agents do most of the implementation. Long documents and long files
waste context, hide rules and invite partial reads. Agents navigate best from
a single entry point that routes to small, focused documents.

## Decision

- **Entry point:** the root `AGENTS.md` holds the rules and a directory map,
  and links each directory's `README.md`. That README explains what the
  directory does and how to contribute to it.
- **Derived docs:** `docs/adr/` records the decisions; `docs/tasks/` holds one
  file per development phase, each with a Definition of Done. Both derive
  from `PROPOSAL.md`, the PRD.
- **Limits:**
  - agentic docs (`*.md`) must be under **100 lines**;
  - code (Go including tests, shell, YAML, Quarto/Typst, configs) must be
    under **300 lines**;
  - aim well under, at about 80 and 250 lines.
- `PROPOSAL.md` is the only exemption. It is the frozen source spec, and the
  ADRs and tasks are its agent-sized derivatives.
- `scripts/check-lines.sh` enforces the limits locally and in CI.
- When a file approaches its limit, **split it by concern**. For example,
  `SKILL.md` moves detail into `skills/caddie/references/`, and a Go file is
  split by type or subcommand.

## Consequences

- Agents can load exactly the context a task needs.
- More files and more links, so every new directory or doc must be linked
  from `AGENTS.md` or its parent README.
- Changing the limits requires superseding this ADR.
