# Development tasks

One file is one development phase. Phases are sequential: each builds on the
previous one and ends in something demonstrable. They are derived from
[PROPOSAL.md §15](../../PROPOSAL.md#15-milestones) (milestones M1–M5), with
M2, M3 and M4 each split in two.

| # | Phase | PRD milestone | Status |
|---|-------|---------------|--------|
| 01 | [Skeleton](01-skeleton.md) | M1 | Done |
| 02 | [Discover: job boards](02-discover-boards.md) | M2 | Done |
| 03 | [Discover: organisation portals](03-discover-portals.md) | M2 | Todo |
| 04 | [Verify, keywords, answers](04-verify-and-answers.md) | M3 | Todo |
| 05 | [Render](05-render.md) | M3 | Todo |
| 06 | [Browser sandbox](06-browser-sandbox.md) | M4 | Todo |
| 07 | [Apply and archive](07-apply-and-archive.md) | M4 | Todo |
| 08 | [Follow-through](08-follow-through.md) | M5 | Todo |

Status values: `Todo` → `In progress` → `Done`. Update them in this table and
in the phase file.

## Phase file format

Every phase file has these sections, in order:

1. **Goal**: one paragraph describing what works at the end.
2. **Inputs**: the PRD sections and ADRs to read first.
3. **Scope** and **Out of scope**.
4. **Deliverables**: the files and commands that will exist.
5. **Definition of Done**: checkboxes. A phase is `Done` only when every box
   is ticked.

Every phase's Definition of Done also implicitly includes these items:

- `scripts/check-lines.sh` passes. Docs stay under 100 lines and code under
  300 lines.
- `go vet ./...` and `go test ./...` pass in CI, with no network access in
  tests.
- `golangci-lint` passes: every function's cyclomatic and cognitive
  complexity is ≤ 15.
- Every new directory has a README linked from `AGENTS.md`.
- The relevant `skills/caddie` section is updated, and every command it shows
  works exactly as written.

## Working a phase

- Work in phase order. If you find work that belongs to a later phase, add a
  checkbox to that phase instead of doing it now.
- If the scope must change, edit the phase file in the same change and say
  why in the commit message.
- If a decision is not covered by an ADR, write one ([how](../adr/README.md)).
