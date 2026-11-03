# 0031. Portal detail_url: an optional trailing `{slug}`

- **Status:** Accepted; amends the `detail_url` matching of [0019](0019-portal-config-schema.md)
- **Date:** 2026-11-03
- **PRD refs:** §7, §11, D11

## Context

ADR 0019 matches a job URL against `detail_url`, whose `{id}` captures
exactly one path segment or query value; the whole path and query must
match. Some portals link job pages with a decorative segment after the id:
BCG's Phenom site links `/global/en/job/<jobId>/<title-slug>`, while
`/global/en/job/<jobId>` serves the same page. `caddie posting save` of
the link a user actually copies was therefore not recognised as
`portal:bcg` (no `ats`, a `url:` ID instead of the `portal:bcg:<jobId>`
that search gives the same job). Loosening matching for every portal is
not an option: ADB's test asserts `/careers/<id>/extra` does not match,
and an over-eager match would assign the wrong portal or ID.

## Decision

- `detail_url` may end in `/{slug}`, opt-in per portal. When matching,
  it matches **one optional path segment** (any text without `/`, `?` or
  `#`), so both `…/job/<id>` and `…/job/<id>/<slug>` match and yield the
  same `{id}`; a second extra segment or a query does not. A `#fragment`
  is ignored as before.
- When search builds a job URL from an id, `/{slug}` is dropped: the URL
  is `…/job/<id>`, so search results and `posting save` agree on the ID,
  and no slug is invented.
- Validation (`<file>: detail_url: …`, exit 3): `{slug}` may appear only
  once, as the last path segment (`/{slug}` at the very end, so after
  `{id}`), and not in a `detail_url` with a query. Fixture:
  `testdata/portals/invalid/detail-url-slug-not-last.yml`.
- A placeholder was chosen over a boolean key (`detail_slug: true`): it
  shows in the pattern itself where the extra segment goes, needs no
  extra schema key, and a portal without it behaves exactly as before.

## Consequences

- BCG job links copied from the site are saved as `portal:bcg:<jobId>`
  with `ats: phenom`, matching the search ID, so search memory and
  `posting save` dedupe correctly.
- Other portals are unchanged; `/{slug}` must be added deliberately.
- Only one trailing segment is accepted. A portal with a deeper or
  mid-path decoration needs a new ADR rather than a looser pattern.
- The slug's content is never checked: any segment matches, so the ID
  alone identifies the job.
