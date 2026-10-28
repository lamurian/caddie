# 0006. Portal adapters and the manual checklist

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §6.4, §7, §12.3, D11, D12

## Context

WHO, UN, World Bank, ADB, IsDB, McKinsey, BCG and Bain run custom portals
(Taleo/Oracle, SuccessFactors, Workday or bespoke ones) with mandatory
accounts, multi-page forms and often CAPTCHAs. Automating them is fragile,
and sometimes not allowed.

## Decision

- **Discovery** uses one generic, config-driven `portal` adapter with one
  tracked YAML file per organisation in `portals/`. Each file gives the
  listing URL, an item selector or JSON path, field mappings and a detail URL
  pattern.
- **Application** to these portals is expected to end in the **manual
  checklist**. This is a designed outcome, not a failure.
- Routing to manual writes a `## Manual steps` section of `- [ ]` items into
  the posting, keeping the agent's understanding of the form actionable.
  Then `caddie move <slug> manual --reason "<one line>"` runs.
- `caddie follow-ups` surfaces manual postings with unchecked items.

## Consequences

- Adding an organisation means a config file, a fixture and a test, with no
  new Go code unless the schema genuinely can't express it.
- Selectors break when portals change, and fixture tests make that visible.
- The user does the final application on these portals by hand, guided by
  the checklist.
