# .github/

CI and dependency automation ([ADR 0013](../docs/adr/0013-dev-tooling.md)).
The Renovate config itself lives at the repo root in
[`renovate.json`](../renovate.json), where Renovate looks first.

```
.github/
└── workflows/ci.yml    # jobs `check` (mise run check) and `sandbox`
```

## CI

- **`check`** installs the tools pinned in `mise.toml` with
  `jdx/mise-action` and runs `mise run check`, exactly what the pre-push
  hook runs. `main` must require this job: Renovate automerges on it.
- **`sandbox`** proves the browser sandbox live (phase 06,
  [ADR 0023](../docs/adr/0023-doctor-browser-stack.md)). It builds
  `caddie`, copies `preferences.example.yml` to `preferences.yml` and the
  fictional `testdata/history/*.md` (not `README.md`) into `history/` on
  the ephemeral runner, never committed. It then runs
  `docker compose -f docker/compose.yml up -d --build --wait` and
  `./caddie doctor`, whose exit code is the verdict (stack up, proxy denies
  `example.com` and allows `careers.who.int`, no published port). On
  failure it dumps `docker compose logs`; it always runs `down -v`. It
  needs outbound network, and an upstream outage of `careers.who.int`
  turns it red.

## Renovate

- Updates mise tools, Go modules, GitHub Actions (digest-pinned) and compose
  images, weekly, 3 days after release. Vulnerability fixes go out at once.
- Every update automerges once `check` passes. The Dependency Dashboard
  issue lists pending and merged updates.
- `renovate.json` is strict JSON (no comments); put notes in `description`.
- Validate edits from the repo root with
  `npx --package renovate -- renovate-config-validator`.

## Contributing

- **Add a check** as a mise task in `mise.toml` and hook it into `check`;
  never add raw steps to `check`, so CI and hooks stay identical. `sandbox`
  is the exception: it needs Docker and the network, which hooks must not.
- Pin every action to a full commit SHA with a version comment.
- Run `actionlint` (or `mise run lint:actions`) after editing a workflow.
