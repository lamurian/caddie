# testdata/documents/

**Fictional** tailored drafts (resumes and cover letters) that the
`document` tests verify and score against `testdata/history/`,
`testdata/postings/` and fictional answers.

| File | Scenario |
|------|----------|
| `resume-traceable.md` | Every claim kind, written fuzzily ("June 2022", "40000", "3"); passes |
| `resume-invented-employer.md` | An invented employer in an entry and a recombined name in prose |
| `resume-stretched-dates.md` | An earlier start date; a month where the history has only a year |
| `resume-upgraded-title.md` | "Senior Data Analyst"; a title moved onto another employer |
| `resume-invented-skill.md` | PyTorch and Kubernetes in prose and in the skills list |
| `cover-letter-answers.md` | Passes only with its answers (a named taskforce, a start date) |

## Contributing

Name each file after the scenario it exercises (`resume-invented-employer.md`),
add it in the same change as its test, and keep every claim fictional.
