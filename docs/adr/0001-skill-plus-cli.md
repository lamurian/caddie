# 0001. Skill + single Go CLI; judgment vs mechanics

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §1, §5, D1, D2, D3

## Context

Caddie must behave identically whether Pi, Goose or another agent drives it,
locally or remotely through Paseo. A Pi extension would only work with Pi.
LLMs drift when tailoring documents, and that drift is dangerous at the risky
points: claims, state transitions and submission.

## Decision

- Caddie ships as **an Agent Skill (`skills/caddie/SKILL.md`) plus one static
  Go binary (`caddie`)**, not as an agent extension.
- **Judgment lives in the skill:** ranking, tailoring, writing and deciding
  what to ask.
- **Mechanics live in the binary:** parsing, verification, keyword coverage,
  memory, state transitions, file moves, rendering and browser argv
  construction.
- The binary makes no LLM calls. It reads `.env` and environment variables.
- A Pi extension may be added later as sugar on top of the CLI (phase 08),
  never instead of it.

## Consequences

- Any agent with a shell tool can use Caddie, and behaviour at the risky
  points is enforced in code and covered by tests.
- The CLI's output is the agent's input, so it must be short, structured
  (`--json`) and stable. Exit codes are part of the contract.
- Rule of thumb for contributors: if two agents could do it differently and
  that matters, it belongs in the binary.
