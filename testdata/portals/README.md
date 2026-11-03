# testdata/portals/

Portal configs and saved portal responses, replayed by the `internal/source`
portal tests through a fake HTTP client
([ADR 0019](../../docs/adr/0019-portal-config-schema.md)).
**Everything here is fictional**: organisations, IDs, hosts and text.

| Path | Holds |
|------|-------|
| `example-html.yml` | A selector-based config (HTML listing, `detail_url` with a query `{id}`) |
| `example-json.yml` | A JSON-path config (POST search API with a body) |
| `example-html/listing.html`, `detail-1001.html` | The listing page `example-html.yml` selects on, and one job page |
| `example-json/listing.json` | `example-json.yml`'s saved response: nested fields, numeric IDs, an untitled item |
| `invalid/*.yml` | Configs that fail validation, one problem each, named after it |
| `who/listing.json` | `portals/who.yml`'s Taleo search response: positional `column[]`, a JSON-array-in-string location |
| `un/listing.json` | `portals/un.yml`'s API response (shape third-party verified, see its notes): `dutyStation[]`, a UTC `endDate` converted by `deadline_tz`, a null `endDate` |
| `adb/listing.html` | `portals/adb.yml`'s Drupal listing: one table per group, suffixed header ids, an untitled row, an empty location, a closing date without `<time>` |
| `isdb/listing.html` | `portals/isdb.yml`'s SuccessFactors search page: titles repeated in desktop/tablet/mobile sections, a relative `data-url`, a tile without Location |
| `bcg/listing.json` | `portals/bcg.yml`'s Phenom widgets response: `refineSearch.data.jobs[]`, an item without `location`, an untitled item |
| `worldbank/listing.html` | `portals/worldbank.yml`'s Cornerstone page: an empty JavaScript shell (`div#cs-root`), why the config is `js_rendered` |
| `bain/detail-70101.html` | `portals/bain.yml`'s job page (discovery-only, no listing): an empty React mount plus the `JobPosting` JSON-LD `posting save` reads |

`portals/mckinsey.yml` has no fixture: its terms forbid automated
collection, so nothing was recorded; its test checks the skip and URL match.

The real configs live in [`portals/`](../../portals/README.md); their
fixtures go in `<org>/` here (`listing.html` or `listing.json`, plus
`detail-<id>.html` for the `posting save` tests); this table covers them,
so the sub-directories need no README.

## Contributing

- Save a real listing once, then replace every name, ID and text with
  fictional values, keeping the markup or JSON shape the config selects on.
  Trim it to a few items, including one with a missing optional field.
- A new validation rule gets an `invalid/<problem>.yml` and a case in
  `portal_config_test.go`.
- Add the fixture in the same change as the test that replays it.
