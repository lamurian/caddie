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

Pre-alpha. Discovery, verification and rendering (phases 01–05) are done;
browser sandbox, apply and follow-ups (06–08) are in progress
([`docs/tasks/`](docs/tasks/README.md)).

## Setup

```bash
mise install                   # pinned Go, Quarto, linters + git hooks
cp .env.example .env
cp preferences.example.yml preferences.yml
cp answers.example.yml answers.yml
go install ./cmd/caddie        # needs $(go env GOPATH)/bin on your PATH
caddie doctor                  # exit 0 is healthy; warn lines are fine
```

Then edit the git-ignored copies and your history:

- **`preferences.yml`**: `contact`, `salary`, `work_mode`, `locations`,
  `roles` (search terms), `sources` (boards, the `portals` list), `ranking.top_n`.
- **`.env`**: API keys, e.g. `CADDIE_ADZUNA_APP_ID`/`_APP_KEY` (else Adzuna is skipped).
- **`answers.yml`**: reusable form answers; the agent adds more as you answer.
- **`history/`**: the only source of facts for every resume and letter:
  `about.md` plus one file per section ([schema](history/README.md)).

Quarto (PDFs) comes with `mise install`. Applying also needs Docker: run
`mise run stack:up` for the [browser stack](docker/README.md), then `caddie doctor`.

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
caddie search --source portal:who --limit 20
caddie seen list --rejected
caddie follow-ups
```

## Documentation

- [`PROPOSAL.md`](PROPOSAL.md): the full product requirements document.
- [`AGENTS.md`](AGENTS.md): contributor entry point, rules, directory map.
- [`docs/adr/`](docs/adr/README.md): architecture decisions.
- [`docs/tasks/`](docs/tasks/README.md): development phases.

## Licence

MIT. See [LICENSE](LICENSE).
