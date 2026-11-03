# 0023. Doctor checks the browser stack, and a stopped stack fails

- **Status:** Accepted
- **Date:** 2026-11-03
- **PRD refs:** §9, §10, §13, D4–D7

## Context

Phase 06 asks `caddie doctor` to prove the sandbox of ADR 0002: the
compose stack is up, the proxy denies a host off the allowlist and allows
one on it, and no port is published. Doctor so far warns when a tool is
missing and fails when a present tool is broken (`quarto`). A stopped stack
sits between the two, and the probes need both a verdict that TLS
interception or the site itself cannot sway and a host that is reliably
listed.

## Decision

- **Scope.** Stack checks run only when `CADDIE_BROWSER_MODE=compose`. In
  `local` mode doctor adds one `browser-stack` warning (agent-browser runs
  outside the sandbox) and skips them. Missing `docker` stays a `warn`.
- **Calls.** Doctor runs `docker compose -f <CADDIE_COMPOSE_FILE> …`
  through `browse.Exec` (injected via `Options.Browser`), each call under
  a 45 s timeout:
  - `config --format json`: the services, and any declared `ports`;
  - `ps --format json`: state and `Publishers` per container (one object
    per line, or one array on older Compose);
  - `exec -T <CADDIE_COMPOSE_SERVICE> curl -sS -o /dev/null --max-time 30
    -w %{http_connect} <url>`: the proxy's answer to CONNECT.
- **Checks.**
  - `browser-stack` fails unless every service in the compose file is
    `running`, naming the stopped ones and `up -d --build`.
  - `ports` fails if any service declares `ports:` or any container
    publishes a host port. The declared half runs even with the stack down.
  - `proxy-deny` wants CONNECT `403` for `https://example.com`.
  - `proxy-allow` wants `200` for `https://careers.who.int/`, the first
    portal host in `portals/README.md`.
  - Any other code, or none (`000`: proxy unreachable), fails. The probes
    run only when `browser-stack` is ok.
- **A stopped stack fails** when the mode is `compose` and `docker` is on
  PATH. The user has said the browser runs in the stack and can start it,
  `caddie browse` would exit 8 without it, and the failure prints the exact
  command. CI then proves the sandbox by exit code alone, without parsing
  the report.

## Consequences

- With Docker installed, a new session shows a `fail` until the user starts
  the stack once; `restart: unless-stopped` keeps it up after that.
- `proxy-allow` needs the network and `careers.who.int` reachable through
  the proxy; an upstream outage reads as a proxy failure (e.g. `503`).
- If `careers.who.int` ever leaves the allowlist, `proxy-allow` must move to
  another listed host.
- Tests fake `browse.Exec`; only a live run (CI, T5) exercises real docker.
