# docker/browser/

The image for the `browser` service: [agent-browser](https://github.com/vercel-labs/agent-browser)
driving Chrome for Testing, headless
([ADR 0002](../../docs/adr/0002-sandboxed-browser.md)).

| Pinned | Where |
|--------|-------|
| `node:24.22.0-bookworm-slim` + digest | `FROM` line |
| agent-browser `0.39.2` | `ARG AGENT_BROWSER_VERSION` |
| Chrome for Testing `156.0.8078.80` + zip sha256 | `ARG CHROME_VERSION`, `ARG CHROME_SHA256` |

- Chrome is unpacked to `/opt/chrome` and selected with
  `AGENT_BROWSER_EXECUTABLE_PATH`. It is not fetched by `agent-browser
  install`, which always takes the newest build and puts it inside
  `~/.agent-browser`, the volume.
- The image runs as `node` (uid 1000). `/home/node/.agent-browser` is
  created 0700 and owned by `node`, so the named volume mounted there
  inherits that owner and mode on first use.
- `curl` is installed so `caddie doctor` can test the proxy from inside.
- `libnss3-tools` lets agent-browser's `--ca-cert` add a CA to Chrome (for
  a network that re-signs TLS).
- The container idles with `sleep infinity`; work arrives through
  `docker compose exec`.

## Contributing

- Bump a version: change the `ARG` (for Chrome, both version and sha256 of
  `…/chrome-for-testing-public/<version>/linux64/chrome-linux64.zip`), then
  `docker compose -f docker/compose.yml up -d --build browser` and run
  `caddie doctor`.
- Bump the base image by tag **and** digest; never use `latest`.
- No stealth or bot-evasion packages, extensions or flags (PRD §10).
