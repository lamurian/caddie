# testdata/sources/

Recorded job board responses, replayed by the `internal/source` tests and
the `cmd/caddie` search and `posting save` tests through a fake HTTP client.
**Everything here is fictional**: companies, people, IDs and URLs.

| File | Recorded from (shape only) |
|------|----------------------------|
| `adzuna-nl.json` | Adzuna `/v1/api/jobs/nl/search/1`: one stated, one predicted salary |
| `remotive.json` | Remotive `/api/remote-jobs`: text salary, Europe and Worldwide jobs |
| `arbeitnow.json` | Arbeitnow `/api/job-board-api`: hybrid, remote and off-topic jobs |
| `greenhouse-acme.json` | Greenhouse `/v1/boards/acme/jobs?content=true` |
| `greenhouse-acme-123.json` | Greenhouse single job with `pay_input_ranges` (used by the golden posting) |
| `lever-globex.json` | Lever `/v0/postings/globex?mode=json` |
| `lever-globex-posting.json` | Lever single posting with `salaryRange` and `lists` |
| `page-jsonld.html` | A careers page with schema.org `JobPosting` JSON-LD in an `@graph` |
| `page-plain.html` | A careers page without JSON-LD: bold and colon headings |

## Contributing

- Record a real response once, then replace every name, ID, URL and text
  with fictional values and trim it to the fields the adapter reads plus a
  few it ignores. Never commit an API key: Adzuna puts it in the URL.
- Keep each file to the smallest response that covers the case, and add it
  in the same change as the test that replays it.
- A board that changes its response shape gets a new fixture next to the
  old one, so both shapes stay tested.
