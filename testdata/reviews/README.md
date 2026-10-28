# testdata/reviews/

Fictional company reviews in the format of PRD §12.4
(`reviews/<company>.md`: frontmatter with `company` and `updated`, then
bullets). `caddie posting save` copies the body into a posting's
`## Reviews` section.

- `acme.md`: the review merged into the golden posting
  `testdata/golden/posting-senior-data-scientist-at-acme.md`.

## Contributing

Name the file after the company slug, as the real `reviews/` directory
does, and keep the content fictional. Add it with the test that reads it.
