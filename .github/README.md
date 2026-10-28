# .github/

CI and dependency automation ([ADR 0013](../docs/adr/0013-dev-tooling.md)).

```
.github/
├── renovate.json5      # Renovate: automatic dependency updates
└── workflows/ci.yml    # the `check` job: mise install + mise run check
```

## CI

The single `check` job installs the tools pinned in `mise.toml` with
`jdx/mise-action` and runs `mise run check`, exactly what the pre-push hook
runs. `main` must require this job: Renovate automerges on it.

## Renovate

- Updates mise tools, Go modules, GitHub Actions (digest-pinned) and compose
  images, weekly, 3 days after release. Vulnerability fixes go out at once.
- Every update automerges once `check` passes. The Dependency Dashboard
  issue lists pending and merged updates.
- Validate edits with
  `npx --package renovate -- renovate-config-validator .github/renovate.json5`.

## Contributing

- **Add a check** as a mise task in `mise.toml` and hook it into `check`;
  never add raw steps to the workflow, so CI and hooks stay identical.
- Pin every action to a full commit SHA with a version comment.
- Run `actionlint` (or `mise run lint:actions`) after editing a workflow.
