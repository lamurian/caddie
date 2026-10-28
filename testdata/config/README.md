# testdata/config/

Fictional configuration fixtures. Files are named `*.fixture.yml` because the
root `.gitignore` ignores `preferences.yml` at any depth.

- `preferences.fixture.yml`: a complete, valid `preferences.yml` with
  placeholder contact details. Used by the `profile` and `doctor` tests and
  the golden file. Run against it with
  `CADDIE_WORKSPACE=testdata CADDIE_PREFERENCES=config/preferences.fixture.yml`.

## Contributing

Keep every value fictional (`example.org`, zeroed phone numbers). Prefer
inline YAML in a test for one-off broken cases; add a file here only when
several tests share it.
