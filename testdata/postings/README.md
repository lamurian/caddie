# testdata/postings/

Fictional saved postings: Markdown with the YAML frontmatter of PRD §12.3.

- `senior-data-scientist-at-acme.md`: a complete, valid posting in status
  `saved`, used by the `workspace` frontmatter tests.
- `lead-biostatistician-at-globex.md`: a `## Requirements` section with
  labels, filler, "X and Y", a single-letter skill, a long phrase and
  keywords the fictional history lacks; used by the `document` keywords
  golden test.

## Contributing

Name each file after its scenario (`posting-no-salary.md`) and add it in
the same change as the test that uses it. Raw HTML or JSON a parser reads
goes here too; keep it to the smallest page that reproduces the structure.
