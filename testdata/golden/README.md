# testdata/golden/

Expected outputs that tests compare byte for byte.

| File | Produced by | Regenerate |
|------|-------------|------------|
| `profile.json` | `caddie profile --json` on `testdata/` | `go test ./cmd/caddie -update` |

## Contributing

Never hand-edit a golden file to make a test pass. Change the code or the
fixture, regenerate with `-update`, and review the diff like code.
