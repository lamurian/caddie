# 0028. Portal pagination: `pages` and `{page}`

- **Status:** Accepted
- **Date:** 2026-11-03
- **PRD refs:** §7, §11, D11

## Context

ADR 0019 allows one GET or POST per portal, with no pagination. The WHO
Taleo job board API answers 25 openings per request whatever page size is
asked for (`pageSize` in the query or body is ignored), so `portals/who.yml`
saw only the 25 newest of about 55 open. Its `pageNo` body field does page:
checked live on 2026-11-03, pages 2 and 3 held the rest and page 4, past the
end, repeated page 3. Other Taleo and search APIs page the same way, by a
page number in the body or the query.

## Decision

- A portal config may set **`pages`**, an integer from 1 to 10: the most
  requests caddie sends to that portal per search. Unset (0) keeps ADR
  0019's single request.
- With `pages`, the placeholder **`{page}`** must appear in `listing_url`
  or `body` (or both). Caddie replaces it with `1`, `2`, … in turn, as
  plain text, so it fits a query value (`?page={page}`), a path segment or
  a JSON number (`"pageNo":{page}`). Pages count from 1; an API that
  counts from 0 or by item offset needs a new ADR.
- `{page}` without `pages`, `pages` without `{page}`, or `pages` outside
  1–10 is an invalid config (exit 3, `<file>: <field>: <problem>`).
- Requests go one after another. Paging stops at `pages`, or at the first
  page that adds no item not already seen (by id, url and title): an empty
  page, or a repeat of an earlier one, as Taleo answers past the end.
  Items already seen are dropped, so the listing stays in page order
  without duplicates.
- A failed page fails the whole portal, as a failed single request does,
  so a partial listing is never reported as complete.
- Headers, method, extraction and IDs are unchanged (ADR 0019).

## Consequences

- WHO lists every opening in a few requests (four today) instead of 25.
- The stop rule needs no per-portal knowledge of totals or page size, but
  costs one extra request when the last page is full, and a portal that
  errors past the end fails at the page after its last; such a portal
  sets `pages` to fit, and its notes say so.
- The cap of 10 keeps a search from crawling a portal; a portal with more
  openings than `pages` pages still lists only the newest, and its notes
  say so.
- The fake-HTTP tests cover paging without network access.
