# docker/proxy/

The browser's egress proxy: tinyproxy with a default-deny host allowlist
([ADR 0002](../../docs/adr/0002-sandboxed-browser.md),
[PRD §9](../../PROPOSAL.md#9-browser-service)).

| File | Holds |
|------|-------|
| `Dockerfile` | Alpine (tag + digest) with tinyproxy 1.11.3, run as user `tinyproxy` |
| `tinyproxy.conf` | Port 8888, `FilterDefaultDeny Yes`, `FilterType ere`, `ConnectPort 443` |
| `allowlist.conf` | One allowed host per line, as an anchored extended regex |

Both config files are copied into the image, so a change needs a rebuild:
`docker compose -f docker/compose.yml up -d --build proxy`.

## Allowlist syntax

tinyproxy matches each line against the request's **host** (not the URL),
case-insensitively. A pattern ends at the first space or `#`; the rest of
the line is a comment.

| Want | Line |
|------|------|
| One exact host | `^jobs\.lever\.co$` |
| A domain and all its subdomains (`*.myworkdayjobs.com`) | `(^\|\.)myworkdayjobs\.com$` |

Always escape dots and anchor both ends: `careers.who.int` unanchored would
also allow `careers.who.int.evil.example`.

The seed list is the PRD §9 hosts plus every portal host (the
[Hosts table](../../portals/README.md#hosts)) and `inspira.un.org`, where UN
applications happen. `example.com` is deliberately absent: `caddie doctor`
uses it to prove the proxy denies.

## Contributing

- Add one host per line with a comment naming the ATS or organisation.
  Prefer exact hosts to wildcards.
- A form often loads scripts or styles from a CDN host; if a page breaks,
  the proxy log (`docker compose -f docker/compose.yml logs proxy`) names
  each refused host ("Proxying refused on filtered domain"). Add only the
  hosts the form needs.
- A new `portals/<org>.yml` needs its hosts here:
  `internal/source/allowlist_test.go` fails otherwise.
- Never allow `example.com`, and never turn `FilterDefaultDeny` off.
