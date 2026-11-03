# 0029. Portal deadline time zone

- **Status:** Accepted
- **Date:** 2026-11-03
- **PRD refs:** §7, §6.5
- **Amends:** the `deadline` rule and schema keys of [0019](0019-portal-config-schema.md)

## Context

ADR 0019 reads a portal's `deadline` as the first unambiguous date in the
field's text. The UN API sends `endDate` as a UTC instant: a job the UN's
own RSS feed shows closing "11/13/2026 23:59:59 PM (New York time)" has
`endDate` `2026-11-14T04:59:59.000Z` (checked 2026-11-03). Taking the
first ten characters gives 14 November, one day after the real closing
date. A deadline that reads a day late is a real harm: follow-ups would
warn too late, and a user could miss the closing date by trusting it.

## Decision

- Portal configs gain an optional key **`deadline_tz`**: an IANA time zone
  name (`America/New_York`). Validation loads it with `time.LoadLocation`
  and refuses an unknown name and `Local` (host-dependent), as
  `deadline_tz: <problem>`; an invalid config is exit 3, as in ADR 0019.
- When `deadline_tz` is set and the whole `deadline` field is an RFC 3339
  timestamp **with an offset** (`Z` or `±hh:mm`), Caddie converts the
  instant to that zone and stores that day as `YYYY-MM-DD`.
- Anything else keeps the ADR 0019 rule: a date-only value, a timestamp
  without an offset (already local), and prose ("Closing: 30 Nov 2026").
  Without `deadline_tz`, a UTC timestamp still gives its UTC date.
- The binary embeds Go's zone database (`time/tzdata`, about 450 KB), so the
  result is the same on every host, including Windows and minimal
  containers without `/usr/share/zoneinfo`.
- `portals/un.yml` sets `deadline_tz: America/New_York`, the zone the UN
  states for its deadlines.

## Consequences

- UN deadlines match the date the job page shows; deadline warnings
  (phase 08) fire on the right day.
- A config author must know the portal's posting zone; the notes record
  the evidence. Portals that send dates without a time are unaffected.
- The binary grows by the size of the zone database.
- Timestamps embedded inside longer text are not converted; no portal
  needs that yet.
