# docker/

The sandboxed browser stack ([ADR 0002](../docs/adr/0002-sandboxed-browser.md),
[PRD §9](../PROPOSAL.md#9-browser-service)), added in phase
[06-browser-sandbox](../docs/tasks/06-browser-sandbox.md).

```
docker/
├── compose.yml            # project "caddie": services browser + proxy
├── browser/Dockerfile     # agent-browser + Chrome for Testing, non-root
└── proxy/                 # tinyproxy: Dockerfile, tinyproxy.conf, allowlist.conf
```

Start it from the repo root, then check it:

```sh
mkdir -p .caddie/trace          # the /trace bind source; see below
docker compose -f docker/compose.yml up -d --build
caddie doctor
```

## Trace directory

Compose never creates `.caddie/trace/` (`up` fails if it is missing), and
the browser's `node` user (uid 1000) must be able to write to it. When your
uid is 1000 (`id -u`), `mkdir -p` is enough. Otherwise grant uid 1000 alone,
so you still own the directory and `caddie move` can empty it:
`setfacl -m u:1000:rwx .caddie/trace`. Avoid `chmod 0777`: any local user
could then plant files that `caddie move` archives. CI, which never moves,
uses `sudo chown 1000:1000` instead. `caddie doctor` fails while the
directory is missing and warns when uid 1000 looks unable to write it.

## Services

- **browser** ([guide](browser/README.md)): agent-browser 0.39.2 driving
  Chrome for Testing in new headless mode, as user `node` (uid 1000). It
  idles (`sleep infinity`) until `caddie browse` runs
  `docker compose exec -T browser agent-browser …`; agent-browser starts
  its own daemon on the first command.
  - `HTTP(S)_PROXY`, `http(s)_proxy` and `AGENT_BROWSER_PROXY` are
    `http://proxy:8888`; `NO_PROXY` is `localhost,127.0.0.1` only.
  - Named volume `browser-state` (`caddie_browser-state`) at
    `/home/node/.agent-browser`, mode 0700: sessions, auth vault, state.
  - The repo's `handoff/` is bind-mounted **read-only** at `/handoff`.
  - `.caddie/trace/` is bind-mounted **read-write** at `/trace`, for trace
    screenshots only ([ADR 0024](../docs/adr/0024-move-archive-and-trace.md)).
- **proxy** ([guide](proxy/README.md)): tinyproxy 1.11 on port 8888, reached
  only on the compose network. It denies every host not matched by
  `proxy/allowlist.conf` (403 "Filtered") and CONNECTs only to port 443.

## Networks

- `sandbox` (`internal: true`): browser and proxy. It has no route out, so
  the browser cannot reach anything but the proxy, even if it ignores the
  proxy settings.
- `egress`: the proxy alone, its way out to allowlisted hosts.

**No service publishes a port**, so there is no CDP or proxy port on the
host.

## Checking it by hand

`caddie doctor` automates these as `browser-stack`, `ports`, `proxy-deny`
and `proxy-allow` ([ADR 0023](../docs/adr/0023-doctor-browser-stack.md)):

```sh
f=docker/compose.yml
# denied: CONNECT answers 403, curl exits 56
docker compose -f $f exec -T browser curl -sS -o /dev/null -w '%{http_connect}' https://example.com
# allowed: CONNECT answers 200
docker compose -f $f exec -T browser curl -sS -o /dev/null -w '%{http_connect}' https://careers.who.int/
# no published ports: prints nothing
docker compose -f $f ps --format '{{.Publishers}}' | grep -v '^\[\]$'
```

## Security invariants (PRD §10)

- The browser volume is mode 0700 and is never bind-mounted into the repo.
- No CDP or other port is exposed on the host.
- No stealth or bot-evasion tooling. A site that blocks headless browsers
  routes to `manual/`.
- Session state and credentials stay in the volume, never in the repo or in
  the agent's context ([ADR 0005](../docs/adr/0005-credentials.md)).
- Chrome runs with `--no-sandbox` (agent-browser's default in a container);
  the container, its non-root user and the proxy are the sandbox.

## Contributing

- **Extending the allowlist:** see [proxy/README.md](proxy/README.md), then
  `docker compose -f docker/compose.yml up -d --build proxy`.
- Pin every version: base images by tag **and** digest, agent-browser and
  Chrome by exact version (Chrome also by sha256). Never use `latest`.
- Never add `ports:`, a bind mount of the browser volume, another writable
  mount besides `/trace`, or another network to `browser`.
- `caddie doctor` must still pass inside the stack (the CI `sandbox` job runs it).
