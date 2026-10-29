# 08. Follow-through

- **Status:** Todo
- **PRD milestone:** M5

## Goal

A morning "follow-ups" shows overdue replies, open manual checklists and
approaching deadlines. All eight organisation portals are configured, and
the scheduled-search scenario is documented.

## Inputs

- PRD §6.5, §12.7 (`apply.*`), §14 (Paseo scenarios), §15 (M5)
- ADRs [0006](../adr/0006-portals-and-manual.md),
  [0011](../adr/0011-posting-state-machine.md),
  [0001](../adr/0001-skill-plus-cli.md)

## Scope

- `caddie follow-ups [--json]` reports three kinds of entry:
  - `applied/` entries past `follow_up_after`;
  - `manual/` entries with unchecked `- [ ]` items;
  - postings in `postings/` or `manual/` whose `deadline` is within
    `apply.deadline_warning_days`.
- Remaining portal configs: `worldbank`, `adb`, `isdb`, `mckinsey`, `bcg`
  and `bain`, each with a fixture and test, and their hosts added to the
  allowlist.
- Skill (`references/follow-up.md`): the follow-ups workflow and the scheduled
  search rule (save the top N, never apply).
- Optional: a thin Pi extension wrapping the CLI. It must not duplicate
  logic. If it is built, record it in an ADR.

## Out of scope

Unattended application, in any form.

## Deliverables

The follow-ups handler and logic in `internal/workspace` or a new package,
six `portals/*.yml` files with fixtures, and the skill reference.

## Definition of Done

- [ ] Follow-ups tests use a fixed clock and cover all three categories,
      plus the empty case.
- [ ] Deadline warnings respect `deadline_warning_days`; past deadlines are
      flagged distinctly.
- [ ] All eight portals are listed in `portals/README.md`, each with a
      passing fixture test (or a documented reason it is discovery-only or
      unsupported).
- [ ] Each PRD §14 scenario has been walked through once with Pi and once
      with Goose, with the results noted in the PR.
- [ ] `portals/un.yml` verified live from a network careers.un.org does not
      block: record a real response, confirm the request body and the
      `data.list`, `jobId`, `postingTitle`, `dutyStation`, `endDate` keys,
      and update the fixture and the notes' ASSUMED section.
- [ ] WHO lists only the 25 newest openings: find a page size or add
      pagination (an ADR 0019 change), or keep it documented.
- [ ] Portal deadlines: decide whether the UN `endDate` (UTC, can be a day
      after the New York closing date) is converted to the posted date.
- [ ] `docs/tasks/README.md` shows every phase as `Done`.
