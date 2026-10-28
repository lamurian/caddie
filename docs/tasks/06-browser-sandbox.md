# 06. Browser sandbox

- **Status:** Todo
- **PRD milestone:** M4 (part 1)

## Goal

A locked-down headless browser runs in Docker behind an allowlisting proxy,
and the agent can drive it only through `caddie browse`, which validates
every call.

## Inputs

- PRD §7 (`browse` rules), §9, §10, §12.8 (browser env vars), §13
- ADRs [0002](../adr/0002-sandboxed-browser.md),
  [0005](../adr/0005-credentials.md)

## Scope

- `docker/compose.yml`:
  - a `browser` service (agent-browser, headless, pinned version), with a
    named volume for `~/.agent-browser` (0700) and `./handoff` mounted
    read-only at `/handoff`;
  - a `proxy` service (tinyproxy or squid), set as `HTTP(S)_PROXY` for the
    browser;
  - no published ports.
- `docker/proxy/allowlist.conf` seeded with the PRD §9 hosts and the portal
  hosts from phase 03.
- `internal/browse`: an argv builder plus an `Exec` interface.
  - `CADDIE_BROWSER_MODE=compose` runs `docker compose exec`; `local` runs
    the binary directly.
  - A timeout is applied.
- `caddie browse <session> <agent-browser args…>` validates that the session
  name is hostname-like and that upload paths are under `/handoff`.
- `caddie doctor` checks that the stack is up, that the proxy blocks a
  non-allowlisted host and allows a listed one, and that no ports are
  published.
- CI job: start the compose stack and run `caddie doctor` inside it.

## Out of scope

The gate, form-filling workflow, `move` and the manual checklist
(phase 07).

## Deliverables

`docker/compose.yml`, `docker/proxy/*`, `internal/browse/*`, the CLI
handler, the CI job and an update to `docker/README.md`.

## Definition of Done

- [ ] Fake-`Exec` tests assert the exact argv for compose and local modes.
- [ ] Invalid session names (`../x`, spaces, empty) are rejected (tested).
- [ ] Uploads outside `/handoff` are rejected, including `..` traversal
      (tested).
- [ ] Doctor proves that the proxy denies `example.com` and allows a listed
      host.
- [ ] No ports are published, and the volume is not bind-mounted into the
      repo.
- [ ] The CI compose job is green.
