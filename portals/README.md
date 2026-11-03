# portals/

Per-organisation configs (tracked, public) for the `portal` source, a
config-driven scraper ([ADR 0006](../docs/adr/0006-portals-and-manual.md)).
Configured: `who`, `un` (phase [03](../docs/tasks/03-discover-portals.md)),
`worldbank`, `adb`, `isdb`, `mckinsey`, `bcg`, `bain` (phase
[08](../docs/tasks/08-follow-through.md)); see Hosts for which are searched.

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
| `headers` | Extra request headers, functional only (`tz`); no credentials or `User-Agent` |
| `item_selector` | CSS selector for one opening in HTML, **or** … |
| `json_path` | dotted path to the array of openings in JSON (`data.items`) |
| `fields` | `id`, `title` (required), `url`, `location`, `deadline`, `salary` |
| `detail_url` | Job page pattern with `{id}`, optionally ending `/{slug}` for one ignored extra segment ([ADR 0031](../docs/adr/0031-portal-detail-slug.md)); required without `fields.url` |
| `pages` | Pages to fetch at most (1–10); `{page}` in `listing_url` or `body` becomes 1, 2…; stops at a page with nothing new ([ADR 0028](../docs/adr/0028-portal-pagination.md)) |
| `page_size` | Items per page (1–1000), for an API paged by item offset: `{offset}` (instead of `{page}`) becomes 0, `page_size`, 2×`page_size`… ([ADR 0030](../docs/adr/0030-portal-offset-paging.md)) |
| `ats` | ATS label: taleo, oracle, successfactors, workday, inspira, cornerstone, phenom, avature, bespoke (required) |
| `js_rendered` | `true` if the listing needs JavaScript: skipped with a notice |
| `skip_reason` | Any other reason search skips the portal (an API refusal, the terms); the notice quotes it |
| `deadline_tz` | IANA zone (`America/New_York`): an RFC 3339 deadline with an offset becomes this zone's date ([ADR 0029](../docs/adr/0029-portal-deadline-timezone.md)) |
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

Enabled in `preferences.yml` → `sources.portals`; `--source portal` queries
every enabled portal, `--source portal:<org>` one. `posting save`
recognises every config here, enabled or not, and takes `ats` from it.

Each item becomes a job with ID `portal:<org>:<id>` (`fields.id`, else the
`{id}` `detail_url` captures from the URL, else a URL hash). `roles.include`
matches the title only. `deadline` text may hold other words ("Closing: 30
Nov 2026"): the first `2026-11-30`, `30 Nov 2026`, `30 November 2026`, `Nov
30, 2026` or `30-Nov-2026` in it is used; `30/11/2026` is ambiguous and
ignored. `salary` counts only with a currency or period ("USD 90,000",
"90,000 per year"), so a grade such as `P-3` is not a salary.

## Hosts

The hosts each config contacts, all on the browser's
[`allowlist.conf`](../docker/proxy/README.md) with the apply hosts named
below (`internal/source/allowlist_test.go` checks it).

| Org | Hosts | Support | Notes |
|-----|-------|---------|-------|
| `un` | `careers.un.org` | Searched | JSON API (POST), `deadline_tz`; applying is on `inspira.un.org` (login) |
| `who` | `careers.who.int` | Searched | Taleo job board API (POST, `tz` header), paged by `pageNo` (25 per page, ADR 0028); applying needs a Taleo account |
| `adb` | `www.adb.org` | Searched | Server-rendered HTML, all openings on one page; titles keep " / <ref>"; applying is on `aces.adb.org` (login) |
| `isdb` | `careers.isdb.org` | Searched | SuccessFactors HTML `/search/`; first 25 only; no deadline in the listing; URL-hash IDs; applying is on `career55.sapsf.eu` (account) |
| `bcg` | `careers.bcg.com` | Searched | Phenom widgets API (POST), paged by offset 500 at a time (unstable order: a run may miss ~30); applying is on `experiencedtalent.bcg.com`, `studenttalent.bcg.com` or `platinion.avature.net` (account) |
| `worldbank` | `worldbankgroup.csod.com` | Discovery-only | Cornerstone: the list comes from a token-authorised API (401 without it) that ADR 0019 cannot call; applying needs a Cornerstone account |
| `bain` | `www.bain.com` | Discovery-only | The listing API refuses direct access (403), so search skips it; `posting save` reads job pages' JSON-LD |
| `mckinsey` | `www.mckinsey.com` | Discovery-only | Terms forbid automated collection, so search calls nothing; applying is on `mckinsey.avature.net` (account); `gateway.mckinsey.com` is the page's own search XHR |

Discovery-only portals set `js_rendered: true` (notice: "listing needs
JavaScript") or `skip_reason`, which the notice quotes: bain's API refusal,
mckinsey's terms. Browse `listing_url` by hand; `posting save` still
recognises their job pages.

## Discovery, not application

These portals usually need accounts, multi-page forms and CAPTCHAs, so
applying normally ends in the **manual checklist**: the designed outcome.

## Contributing a portal

1. Save a representative listing page (and one detail page) as fixtures in
   `testdata/portals/<org>/`. Strip anything personal.
2. Write `portals/<org>.yml` against the schema.
3. Add a test that parses the fixture into the expected postings.
4. List its hosts above and add them to `docker/proxy/allowlist.conf`.
5. Only scrape sites whose terms allow automated access (PRD §3).
