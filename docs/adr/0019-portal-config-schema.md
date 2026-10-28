# 0019. Portal config schema, IDs and enabling

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §7, §11, D11

## Context

ADR 0006 chose one config-driven `portal` adapter with a YAML file per
organisation, but left the schema open: how items and fields are located,
how jobs are identified, how portals are enabled, and how far the scraper
goes (JavaScript, logins, pagination). Adding an organisation must need
only a YAML file, a fixture and a test.

## Decision

- **Schema.** `portals/<org>.yml`, `<org>` matching `[a-z0-9-]+`, decoded
  strictly (unknown keys are errors). Keys: `name` (used as the company),
  `listing_url`, `method` (`GET` default, or `POST`), `body` (POST only),
  exactly one of `item_selector` (HTML) or `json_path` (JSON), `fields`
  (`id`, `title`, `url`, `location`, `deadline`, `salary`), `detail_url`
  (a pattern with `{id}`), `ats` (a label: taleo, oracle, successfactors,
  workday, inspira, bespoke…), `js_rendered` and `notes`. `name`,
  `listing_url` and `ats` are always required; otherwise `fields.title` is,
  and `fields.url` unless `fields.id` and `detail_url` are both set.
- **HTML** items and fields are CSS selectors, parsed with
  `github.com/andybalholm/cascadia` on top of `golang.org/x/net/html` (the
  parser ADR 0017 chose), never regular expressions. A field selector is
  relative to the item and may end in `@attr` to read an attribute; an
  empty selector before `@` is the item itself (`@data-id`).
- **JSON** uses an in-house dotted subset: keys separated by `.`;
  `json_path` ends at an array of objects and a field path is read inside
  one item (`location.city`). No filters, wildcards or indices: portal APIs
  do not need them, so a full JSONPath library is not worth a dependency.
- **IDs** extend ADR 0016: `portal:<org>:<native id>`, where the native ID
  is `fields.id`, else the `{id}` that `detail_url` captures from the job
  URL, else the 12-hex-digit URL hash used for `url:` IDs.
  `Job.Source` is `portal:<org>`.
- **Enabling.** `preferences.yml` `sources.portals` lists the orgs to
  query. `--source portal` means every enabled portal and
  `--source portal:<org>` one. A listed org without a file, or a
  `js_rendered: true` config, is a notice, not an error. An invalid config
  is exit 3 with `<file>: <field>: <problem>`.
- **Scope of a fetch.** One GET or POST to `listing_url` per portal: no
  JavaScript rendering, no authentication and no pagination beyond what
  that one response returns.
- **`posting save`** sets `ats` from the config and `ats_host` from the URL
  host for a URL that a portal's `detail_url` (or, without one, its listing
  host) matches.
- `Portal.Hosts()` lists each config's hosts to seed the phase 06 egress
  allowlist.

## Consequences

- A new organisation is a YAML file, a fixture and a test; Go changes only
  when the schema cannot express a portal, which then needs a new ADR.
- Validation errors name the file and key, so a broken config is fixed
  without reading Go.
- Portals that render listings with JavaScript, need a login or paginate
  are covered only partly or not at all; their configs say so in `notes`,
  and the user searches them by hand.
- Selectors break when a portal's markup changes; fixture tests catch that
  in CI only for the saved page, so a live change shows up as an empty
  result or a notice.
