# testdata/documents/

**Fictional** tailored drafts (resumes and cover letters) that the
`document` tests verify and score against `testdata/history/`,
`testdata/postings/` and fictional answers.

Each follows the draft format in
[`tailor.md`](../../skills/caddie/references/tailor.md) and
[`cover-letter.md`](../../skills/caddie/references/cover-letter.md): no
frontmatter, title, name, contact details or links (the render template
adds them). A resume starts at `## About`; a letter at its salutation.

| File | Scenario |
|------|----------|
| `resume-traceable.md` | Every claim kind, written fuzzily ("June 2022", "40000", "3"); passes |
| `resume-invented-employer.md` | An invented employer in an entry and a recombined name in prose |
| `resume-stretched-dates.md` | An earlier start date; a month where the history has only a year |
| `resume-upgraded-title.md` | "Senior Data Analyst"; a title moved onto another employer |
| `resume-invented-skill.md` | PyTorch and Kubernetes in prose and in the skills list |
| `cover-letter-answers.md` | Passes only with its answers (a named taskforce, a start date) |
| `keywords-draft-lead-biostatistician.md` | Resume draft for `postings/lead-biostatistician-at-globex.md`: covers most keywords, omits some the history supports, adds an unbacked one (SAS), which is also its only verify failure |

## Contributing

Name each file after the scenario it exercises (`resume-invented-employer.md`),
add it in the same change as its test, and keep every claim fictional.
