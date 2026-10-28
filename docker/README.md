# docker/

The sandboxed browser stack ([ADR 0002](../docs/adr/0002-sandboxed-browser.md),
[PRD §9](../PROPOSAL.md#9-browser-service)). It is added in phase
[06-browser-sandbox](../docs/tasks/06-browser-sandbox.md).

```
docker/
├── compose.yml
└── proxy/allowlist.conf
```

## Services

- **browser**: the agent-browser daemon, headless (Chrome new headless mode).
  - `HTTP(S)_PROXY` points to `proxy`.
  - A named volume holds `~/.agent-browser`.
  - `./handoff` is mounted read-only at `/handoff`.
  - **No published ports.**
- **proxy**: tinyproxy (or squid) with a domain allowlist. Requests to any
  domain not on the list fail.

`caddie browse <session> …` runs `docker compose exec browser agent-browser …`.
The agent never talks to the container directly.

## Security invariants (PRD §10)

- The browser volume is mode 0700 and is never bind-mounted into the repo.
- No CDP or other port is exposed on the host.
- No stealth or bot-evasion tooling. A site that blocks headless browsers
  routes to `manual/`.
- Session state and credentials stay in the volume, never in the repo or in
  the agent's context ([ADR 0005](../docs/adr/0005-credentials.md)).

## Contributing

- **Extending the allowlist:** add one hostname per line to
  `proxy/allowlist.conf`, with a comment naming the ATS or organisation.
  Prefer exact hosts to wildcards.
- Pin image versions; never use `latest`.
- `caddie doctor` must still pass inside the stack (CI runs it there).
