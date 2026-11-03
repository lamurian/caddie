# 0025. The skill lives in `.agents/skills/`

- **Status:** Accepted
- **Date:** 2026-11-03
- **PRD refs:** §8, §11, D1
- **Supersedes:** the skill path in [0001](0001-skill-plus-cli.md)

## Context

Caddie is used in place: the user clones the repo, builds `caddie`, fills
in `history/`, then runs their agent inside the clone. The skill was at
`skills/caddie/`, the layout for repos that publish skills for others to
install, so every user had to install or point their agent at it by hand.
Several agents discover project skills in `.agents/skills/` automatically.

## Decision

- The skill moves to `.agents/skills/caddie/` with no other change to its
  content or structure. Nothing remains at `skills/`.
- `.agents/` gets a README (AGENTS.md rule 6); the skill's own README
  stays at `.agents/skills/caddie/README.md`.
- The Go tests that compare the skill against the CLI read the new path.
- ADRs 0001, 0012 and 0014 keep the old path as history; this ADR is the
  current location.

## Consequences

- An agent started in the clone finds the skill with no install step.
- The folder is hidden: `ls` and some file browsers skip it, so
  `AGENTS.md` and the root README must point to it.
- A contributor's coding agent also loads the skill while working on
  Caddie. Its description is about job applications, so it should not
  trigger on development tasks, but contributors should know it is there.
- Agents that look elsewhere (for example `.claude/skills/`) still need a
  symlink or a manual install.
