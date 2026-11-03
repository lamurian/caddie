# Exit codes: what each command returns

Every `caddie` command exits with one of these codes; errors go to stderr
as one line, `caddie <command>: <message>`. Branch on the code, then read
the message. A code not listed for a command is exit 9: report it.

<!-- exit-codes:begin -->
| Code | Meaning | What the agent does |
|------|---------|---------------------|
| 0 | Success | Continue |
| 1 | A check failed (`doctor` found a problem; `verify` found an untraceable claim) | Read the output, fix the cause, rerun |
| 2 | Usage error: unknown command, bad flag or argument | Fix the command line |
| 3 | Configuration error: `.env`, an env var or `preferences.yml` is invalid or missing | Show the message to the user |
| 4 | Malformed input: `history/` or a posting breaks its schema; the message names file and line | Show the file and line to the user |
| 5 | Not found: no job posting at a URL, an unknown slug or a missing file | Check the URL, slug or path |
| 6 | Illegal posting status transition | Do not edit `status` by hand; follow the workflow |
| 7 | Refused by a safety rule: `render` on an unverified or edited draft; `browse` with a bad session name, an upload outside `/handoff` or a screenshot outside `/trace` | Stop; never work around it |
| 8 | External tool or network failure: HTTP, `quarto`, `docker` or agent-browser (whose own error is printed first) | Run `caddie doctor`, retry once, then tell the user |
| 9 | Internal error | Report it to the user |
<!-- exit-codes:end -->

## By command

Every command exits 0 on success, 2 on a bad flag or argument and 3 on an
invalid `.env`; those are left out below unless something else causes them.

- `caddie doctor`: 1 a `fail` line: show it verbatim and stop (`warn` is
  fine).
- `caddie profile`: 3 `preferences.yml` missing or invalid; 4 `history/`
  breaks its schema.
- `caddie search`: 2 unknown `--source` or negative `--limit`; 3
  `preferences.yml` or a `portals/*.yml` is invalid; 8 every source
  failed (one failing source is only a `notices` line).
- `caddie seen add|list`: nothing else; a bad `seen.jsonl` line is a
  `problems` entry, not an error.
- `caddie posting save <url>`: 2 not an http(s) URL; 3 a `portals/*.yml`
  is invalid; 5 the page has no job posting; 8 the page could not be
  fetched: retry once, then give the user the URL.
- `caddie posting steps <slug> "<step>"…`: 2 no step, an empty step or
  a bad slug; 4 the posting is malformed; 5 no posting in `postings/` or
  `manual/`; 6 the posting is already in `applied/`.
- `caddie answers add|show|list`: 2 bad `--key`, `--scope` or slug; 4
  `answers.yml` (file and line given) or the `--for` posting is
  malformed; 5 `--for` names no posting in `postings/`, `manual/` or
  `applied/`.
- `caddie verify <doc.md>`: 1 untraceable claims, one per line; 4
  `history/`, `answers.yml` or the posting is malformed; 5 no such file.
- `caddie keywords <posting.md> <doc.md>`: 0 whatever the coverage; 4
  `history/` is malformed; 5 a file is missing.
- `caddie render resume|cover-letter <slug>`: 2 unknown kind; 3 template
  missing or `preferences.yml` invalid; 4 `history/` or the posting is
  malformed; 5 no draft or no posting; 6 the posting is in `applied/` or
  `manual/`, or its status is past `rendered`; 7 the draft is unverified
  or changed since `caddie verify`: verify again; 8 quarto failed.
- `caddie browse <session> …`: 2 no session or no command; 3 invalid
  browser settings; 7 bad session name, `--session`, `batch`, a sandbox
  flag, an upload outside `/handoff` or a screenshot outside `/trace`:
  fix the call, never work around it; 8 agent-browser or Docker failed
  or timed out (its own error first): if it is about the page, take a
  fresh `snapshot -i` and retry once, else run `caddie doctor`.
- `caddie move <slug> applied|manual`: 2 bad target, `manual` without a
  one-line `--reason`, or `--reason` with `applied`; 3 `preferences.yml`
  invalid; 4 the posting or `history/` is malformed; 5 no posting
  anywhere; 6 `applied` from a status other than `rendered`, the posting
  already archived, or `<dest>/<slug>` exists: tell the user, never
  delete anything; 9 after the JSON: moved, but `handoff/` was not
  cleared: do not move again, tell the user.
- `caddie follow-ups`: 3 `preferences.yml` invalid (a negative
  `apply.deadline_warning_days`); 4 a posting in `applied/`, `manual/` or
  `postings/` is malformed or has a date that is not `YYYY-MM-DD` (file
  named): show it to the user.
