# testdata/

Fixtures for Go tests. **Everything here is fictional.** Never copy real
personal data, real `history/` content, real answers or session state in here.

```
testdata/
├── config/       # fictional preferences (*.fixture.yml: preferences.yml is ignored)
├── history/      # a fictional CV following the history schema
├── postings/     # sample postings (Markdown + frontmatter) and raw HTML/JSON
├── portals/      # saved listing/detail pages, one directory per organisation
├── sources/      # recorded board API responses (adzuna, remotive, …)
└── golden/       # expected outputs: keyword reports, profile JSON, PDF text
```

Each directory has its own README: [config](config/README.md),
[history](history/README.md), [postings](postings/README.md),
[golden](golden/README.md). `portals/` and `sources/` arrive with their
phases and get one then.

## Rules

- Tests never hit the network. Record a response once, then commit it here.
- **Golden files** are regenerated with `go test <package> -update` (a flag
  that each package owning a golden file defines, e.g.
  `go test ./cmd/caddie -update` for `golden/profile.json`) and reviewed like
  code. Never hand-edit golden output
  just to make a test pass.
- The fictional history should exercise edge cases: a `Present` end date,
  multiple bullets, publications with several authors, and an entry with no
  location.
- `document.Verify` needs claims that are exact, fuzzy, false-positive and
  answer-sourced, so the fixture CV should make each kind possible
  ([PRD §13](../PROPOSAL.md#13-testing)).
- Markdown and YAML fixtures fall under the line limits like any other file.
  HTML and JSON fixtures are not checked, but keep them minimal anyway: the
  smallest page that reproduces the structure.

## Contributing

Add the fixture in the same change as the test that uses it, and name it
after the scenario (`posting-no-salary.md`, not `test3.md`).
