# 0030. Portal offset paging: `{offset}` and `page_size`

- **Status:** Accepted
- **Date:** 2026-11-03
- **PRD refs:** §7, §11, D11

## Context

ADR 0028 pages a portal by a page number, `{page}` = 1, 2…, and leaves an
API that counts by item offset to a new ADR. BCG's Phenom widgets API is
one: it caps `size` at 500 and pages by an item offset in the body
(`"from":0`, then 500, …), so `portals/bcg.yml` saw 500 of the 940 open.
Checked live on 2026-11-03 with caddie's own user agent: from 500 returned
the other 440 and from 1000 none. Elasticsearch-style search APIs
(`from`/`size`, `offset`/`limit`, `start`/`rows`) page the same way.

## Decision

- A paginated portal (`pages` set, ADR 0028) may use the placeholder
  **`{offset}`** in `listing_url` or `body` instead of `{page}`, with the
  key **`page_size`**, an integer from 1 to 1000: the items per page the
  request asks for, as written elsewhere in the request (`"size":500`).
- Request *n*, counting from 0, replaces `{offset}` with *n* × `page_size`
  (0, 500, 1000, …) as plain text, as `{page}` is replaced with *n* + 1.
- Invalid configs (exit 3, `<file>: <field>: <problem>`): `{offset}`
  without `pages`; `{offset}` without `page_size`; `page_size` without
  `{offset}` (so it is never set with `{page}`); both `{page}` and
  `{offset}`; `page_size` outside 1–1000. `pages` needs one of the two
  placeholders.
- Everything else is ADR 0028's: requests go one after another, at most
  `pages`, stopping at the first page that adds no new item; items seen
  before are dropped; a failed page fails the whole portal.
- Caddie does not set the size in the request itself: the config writes
  it, and `page_size` must match it. A mismatch skips or repeats items;
  the repeats are dropped, and the fixture test pins the offsets sent.

## Consequences

- BCG lists all its openings in three requests (500, 440, then an empty
  page that stops paging), with `pages: 4` leaving room to 2000.
- `{page}` configs, such as WHO's, are unchanged.
- BCG's default order is not stable between requests, so a run can see
  an opening on both pages and miss another; offset paging cannot fix
  that, and the config's notes say so.
- A portal that counts pages from 0 can use `{offset}` with `page_size: 1`.
