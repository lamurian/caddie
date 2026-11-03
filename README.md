# Caddie

A caddie carries the bag, knows the course and advises on the shot, but the
player takes the swing.

Caddie is an AI-assisted job application workflow: a single Go binary plus an
Agent Skill that any coding agent (Pi, Goose, Claude Code…, driven remotely
through Paseo) can load. It:

- finds postings on EU job boards and organisation portals, ranks them
  against your history and preferences, and remembers what you rejected;
- writes a tailored resume and cover letter whose every claim traces back to
  your `history/`, rendered to PDF with Quarto;
- fills forms in a sandboxed browser and **stops for your approval before
  submitting**; tracks follow-ups and deadlines.

## Status

Alpha. Every development phase (01–08) is done
([`docs/tasks/`](docs/tasks/README.md)): discovery on boards and portals,
verified tailoring and rendering, the sandboxed browser with the submit
gate, archiving and follow-ups.

## Setup

```bash
mise install      # pinned Go, Quarto, linters + git hooks
mise run init     # compile caddie, then a short Q&A that writes and validates your config
```

`mise run init` installs `caddie` into mise's Go (on your PATH wherever mise
is active; `mise run compile` alone rebuilds it), copies
`.env.example` to `.env`, then asks a default set of questions (roles,
salary, work mode, locations, sources, a few reusable form answers) for
whichever of `preferences.yml` and `answers.yml` is missing. Press Enter
to keep a default. Both files are validated against
[`schemas/`](schemas/README.md) and are git-ignored; existing files are
validated, never overwritten.

Prefer to answer in chat? Run `mise run compile` instead, then start your
agent here: it runs the same first-run check (`caddie validate`, then `caddie init --questions`),
asks the default questions plus a few of its own, and writes both files.

Then:

- **`history/`**: the only source of facts for every resume and letter:
  `about.md` plus one file per section ([schema](history/README.md)).
- **`.env`**: API keys, e.g. `CADDIE_ADZUNA_APP_ID`/`_APP_KEY` (else Adzuna is skipped).
- Edit `preferences.yml` or `answers.yml` any time; run `caddie validate` after.

Quarto (PDFs) comes with `mise install`. Applying also needs Docker: run
`mise run stack:up` for the [browser stack](docker/README.md), then `caddie doctor`
(exit 0 is healthy; `warn` lines are fine).

## Using it

Start your agent in this directory; it loads the skill in
[`.agents/skills/caddie/`](.agents/skills/README.md) by itself. Then say:

| You say | What happens |
|---------|--------------|
| "find jobs" | Ranked top N with match reasoning; reply e.g. "keep 1, 4, 7" to save them to `postings/<slug>.md` |
| "tailor &lt;slug&gt;" | Resume draft, checked by `caddie verify` and `caddie keywords`; iterate, then approve to render the PDF |
| "cover letter &lt;slug&gt;" | At most five questions (only unanswered ones), a verified draft, then the PDF |
| "apply &lt;slug&gt;" | Fills the form, posts a screenshot and stops: reply `submit` or `manual` |
| "follow-ups" | Morning check: overdue replies, open manual checklists, near deadlines |

Caddie never submits on its own; only your `submit` does. Login walls and
CAPTCHAs become a checklist in `manual/`. A scheduled search never applies:

```bash
paseo run --provider pi "find jobs, save the top 5 to postings/ without asking, then stop"
```

Organisation portals: WHO, UN, ADB, IsDB and BCG are searched; World Bank,
Bain and McKinsey are discovery-only (browse them yourself; saving a job URL
still works); applying to any portal ends in the manual checklist
([`portals/`](portals/README.md)).

**Files:** `postings/` (saved), `resumes/`, `cover-letters/` (drafts, PDFs;
copies in `handoff/`), `applied/` (submitted + trace), `manual/` (checklists).

## The CLI directly

The agent drives it, but every command also works by hand. Full list and
exit codes: [`cmd/README.md`](cmd/README.md).

```bash
caddie validate
caddie search --source portal:who --limit 20
caddie seen list --rejected
caddie follow-ups
```
