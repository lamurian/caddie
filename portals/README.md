# portals/

Per-organisation configs for the `portal` source adapter, which is a
config-driven scraper ([ADR 0006](../docs/adr/0006-portals-and-manual.md)).
These files are tracked and public.

Planned: `who.yml`, `un.yml`, `worldbank.yml`, `adb.yml`, `isdb.yml`,
`mckinsey.yml`, `bcg.yml`, `bain.yml`. The first two arrive in phase
[03](../docs/tasks/03-discover-portals.md) and the rest in phase
[08](../docs/tasks/08-follow-through.md).

## Config shape

`portals/<org>.yml`, where `<org>` matches `[a-z0-9-]+` and is the name
used in `sources.portals` and `--source portal:<org>`
([ADR 0019](../docs/adr/0019-portal-config-schema.md)). Unknown keys are
errors; `caddie search` exits 3 naming the file and the key.

| Key | Meaning |
|-----|---------|
| `name` | Organisation name, used as the company (required) |
| `listing_url` | Absolute http(s) URL of the listing page or API (required) |
| `method`, `body` | `GET` (default) or `POST`; `body` is the POST's JSON |
| `item_selector` | CSS selector for one opening in HTML, **or** … |
| `json_path` | dotted path to the array of openings in JSON (`data.items`) |
| `fields` | `id`, `title` (required), `url`, `location`, `deadline`, `salary` |
| `detail_url` | Job page pattern with `{id}`; required without `fields.url` |
| `ats` | ATS label: taleo, oracle, successfactors, workday, inspira, bespoke (required) |
| `js_rendered` | `true` if the listing needs JavaScript: skipped with a notice |
| `notes` | Free text: limitations, terms of use |

For HTML a field is a CSS selector inside the item, optionally ending in
`@attr` (`a@href`; `@data-id` is the item's own attribute). For JSON it is a
dotted key path inside the item (`location.city`); a number indexes an array
(`column.0`), and a string holding a JSON array (`"[\"Geneva\"]"`) reads as
that array, joined with ", ".

```yaml
name: Example Health Organization
listing_url: https://careers.example.org/vacancies
item_selector: ul.vacancies li.vacancy
fields:
  id: "@data-id"
  title: a.title
  url: a.title@href   # relative links resolve against listing_url
  deadline: .closing-date
detail_url: https://careers.example.org/job?id={id}
ats: taleo
```

Enabled from `preferences.yml` → `sources.portals`; `caddie search
--source portal` queries every enabled portal, `--source portal:<org>` one.

Each item becomes a job with ID `portal:<org>:<id>` (`fields.id`, else the
`{id}` `detail_url` captures from the URL, else a URL hash). `roles.include`
matches the title only. `deadline` text may hold other words ("Closing: 30
Nov 2026"): the first `2026-11-30`, `30 Nov 2026`, `30 November 2026`, `Nov
30, 2026` or `30-Nov-2026` in it is used; `30/11/2026` is ambiguous and
ignored. `salary` counts only with a currency or period ("USD 90,000",
"90,000 per year"), so a grade such as `P-3` is not a salary.

`caddie posting save` recognises every config here, enabled or not, and
takes `ats` from it.

## Hosts

The hosts each config contacts (`listing_url`, `detail_url`). They seed
`docker/proxy/allowlist.conf` in phase
[06](../docs/tasks/06-browser-sandbox.md).

| Org | Hosts | Notes |
|-----|-------|-------|
| `un` | `careers.un.org` | JSON API (POST); applying is on `inspira.un.org` (login) |

## Discovery, not application

These portals usually need accounts, multi-page forms and CAPTCHAs, so
applying to them normally ends in the **manual checklist**. That is the
designed outcome, not a failure.

## Contributing a portal

1. Save a representative listing page (and one detail page) as fixtures in
   `testdata/portals/<org>/`. Strip anything personal.
2. Write `portals/<org>.yml` against the schema.
3. Add a test that parses the fixture into the expected postings.
4. List its hosts in the table above (`docker/proxy/allowlist.conf` from
   phase 06).
5. Only scrape sites whose terms allow automated access (PRD §3).
