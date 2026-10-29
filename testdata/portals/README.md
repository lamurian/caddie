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
| `un/listing.json` | `portals/un.yml`'s API response (shape partly assumed, see its notes): `dutyStation[]`, a null `endDate` |

The real configs live in [`portals/`](../../portals/README.md); their
fixtures go in `<org>/` here (no README per sub-directory: this table covers them) (`listing.html` or `listing.json`, plus
`detail-<id>.html`, which the `posting save` tests replay).

## Contributing

- Save a real listing once, then replace every name, ID and text with
  fictional values, keeping the markup or JSON shape the config selects on.
  Trim it to a few items, including one with a missing optional field.
- A new validation rule gets an `invalid/<problem>.yml` and a case in
  `portal_config_test.go`.
- Add the fixture in the same change as the test that replays it.
