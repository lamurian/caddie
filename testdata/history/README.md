# testdata/history/

A **fictional** career history used by the `profile` tests and the
`caddie profile` golden file (`testdata/golden/profile.json`). It follows the
schema in [`history/README.md`](../../history/README.md) and covers its edge
cases:

- a `Present` end date and a wrapped (continued) bullet (`employment.md`);
- an entry with no location line (`employment.md`, second entry);
- full month names, `Sept`, hyphen and em-dash separators, year-only ranges
  (`consultancies.md`, `education.md`);
- an entry with no bullets (`education.md`);
- publications with several authors joined by `,`, `and` and `;`, a
  year-only date and a bullet (`publication.md`), and a presentation.

The parser skips this `README.md`, just as it skips `history/README.md`.

Change the fixture and the golden file together: `go test ./cmd/caddie -update`.
