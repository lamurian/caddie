# testdata/portals/

Portal configs and saved portal responses, replayed by the `internal/source`
portal tests through a fake HTTP client
([ADR 0019](../../docs/adr/0019-portal-config-schema.md)).
**Everything here is fictional**: organisations, IDs, hosts and text.

| Path | Holds |
|------|-------|
| `example-html.yml` | A selector-based config (HTML listing, `detail_url` with a query `{id}`) |
| `example-json.yml` | A JSON-path config (POST search API with a body) |
| `example-json/listing.json` | Its saved response: nested fields, numeric IDs, an untitled item |
| `invalid/*.yml` | Configs that fail validation, one problem each, named after it |
| `<org>/` | Saved listing (and one detail) response for `portals/<org>.yml` |

The real configs live in [`portals/`](../../portals/README.md); their
fixtures go in `<org>/` here (`listing.html` or `listing.json`, plus
`detail.html`).

## Contributing

- Save a real listing once, then replace every name, ID and text with
  fictional values, keeping the markup or JSON shape the config selects on.
  Trim it to a few items, including one with a missing optional field.
- A new validation rule gets an `invalid/<problem>.yml` and a case in
  `portal_config_test.go`.
- Add the fixture in the same change as the test that replays it.
