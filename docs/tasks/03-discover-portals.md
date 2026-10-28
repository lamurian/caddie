# 03. Discover: organisation portals

- **Status:** Todo
- **PRD milestone:** M2 (part 2)

## Goal

`caddie search` also returns openings from international organisations'
career portals, through one config-driven `portal` adapter, with the first
two organisations configured.

## Inputs

- PRD §7 (search adapters), §11 (`portals/`), §13
- ADRs [0006](../adr/0006-portals-and-manual.md),
  [0002](../adr/0002-sandboxed-browser.md) (for the allowlist hosts)

## Scope

- Define the portal config schema:
  - `listing_url`;
  - an item CSS selector **or** a JSON path;
  - field mappings for title, url, location, deadline and salary;
  - a detail URL pattern;
  - an `ats` label.
  Document it in `portals/README.md`.
- `internal/source/portal`: load `portals/<org>.yml`, fetch the listing,
  extract items and normalise them to `Job`. Use an HTML parsing library
  (record the choice in the PR), not regexes.
- Configs for the first two organisations: `who.yml` and `un.yml`.
- Enable them through `preferences.yml` `sources.portals`, with
  `--source portal:<org>` filtering.
- `posting save` works for portal URLs and sets `ats` and `ats_host`.

## Out of scope

Applying through portals (that is the manual path in phase 07). The six
remaining configs are phase 08. No JavaScript-rendered scraping: if a portal
needs it, document the limitation in its config and skip it.

## Deliverables

`internal/source/portal*.go`, `portals/who.yml`, `portals/un.yml`,
`testdata/portals/{who,un}/` and a schema section in `portals/README.md`.

## Definition of Done

- [ ] Fixture tests parse each organisation's saved listing into the expected
      jobs (titles, URLs, deadlines).
- [ ] A config that fails validation (missing listing URL, both selector and
      JSON path set) gives a clear error naming the file and the field.
- [ ] Both selector-based and JSON-path-based configs are exercised in tests.
- [ ] `caddie search --source portal:who` works from a recorded response.
- [ ] Portal hosts are listed so that phase 06 can seed the allowlist.
- [ ] Adding an organisation needs only a YAML file, a fixture and a test
      (demonstrated by the second organisation).
