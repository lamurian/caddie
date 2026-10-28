# portals/

Per-organisation configs for the `portal` source adapter, which is a
config-driven scraper ([ADR 0006](../docs/adr/0006-portals-and-manual.md)).
These files are tracked and public.

Planned: `who.yml`, `un.yml`, `worldbank.yml`, `adb.yml`, `isdb.yml`,
`mckinsey.yml`, `bcg.yml`, `bain.yml`. The first two arrive in phase
[03](../docs/tasks/03-discover-portals.md) and the rest in phase
[08](../docs/tasks/08-follow-through.md).

## Config shape

The schema is fixed in phase 03. It covers:

- `listing_url`: the page or API endpoint that lists openings;
- an item selector (CSS) **or** a JSON path;
- field mappings: title, location, deadline, salary, URL;
- a detail URL pattern;
- the expected ATS family (Taleo/Oracle, SuccessFactors, Workday, bespoke).

Enabled from `preferences.yml` → `sources.portals`.

## Discovery, not application

These portals usually need accounts, multi-page forms and CAPTCHAs, so
applying to them normally ends in the **manual checklist**. That is the
designed outcome, not a failure.

## Contributing a portal

1. Save a representative listing page (and one detail page) as fixtures in
   `testdata/portals/<org>/`. Strip anything personal.
2. Write `portals/<org>.yml` against the schema.
3. Add a test that parses the fixture into the expected postings.
4. Add the portal's host to `docker/proxy/allowlist.conf`.
5. Only scrape sites whose terms allow automated access (PRD §3).
