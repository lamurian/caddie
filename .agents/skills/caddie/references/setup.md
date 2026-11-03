# Setup: the first-run check

Run at the start of every session, before any other workflow.

## 1. Check

1. Run `caddie validate`.
   - Exit 0: both files are valid. Go to step 4.
   - Exit 3 with `file not found`: a file is missing. Go to step 2.
   - Exit 3 or 4 otherwise: show the message (file, key or line) to the
     user, fix the file with them, then run `caddie validate` again.
     Never set `apply.require_gate` to `false`.

## 2. Ask

1. Run `caddie init --questions`. It prints JSON: `missing` (the files to
   create) and `questions`, each with `id`, `prompt`, `kind`, `default`,
   `options` and `required`.
2. Ask every question in **one** message, as a numbered list. Show each
   default and its options; say that an empty reply keeps the default.
   - `list` and `choices`: comma-separated; `choices` must come from
     `options`, most preferred first.
   - `int`: a whole number, 0 or more.
3. In the same message, add up to five questions of your own that the
   user's `history/` and target roles make useful, for example:
   - target countries and visa constraints for those countries;
   - which company boards (Greenhouse, Lever slugs) to watch;
   - recurring form answers (`salary_expectation`, `languages`,
     `relocation`, `why_<sector>`).
   Ask for no password, API key or other secret: keys go in `.env`, which
   the user edits by hand.

## 3. Write

1. Write the replies to a temporary JSON file outside the repo, keyed by
   question `id`. Lists are JSON arrays; numbers are numbers; leave out a
   question the user skipped (its default applies). For example:
   `{"roles.include": ["epidemiologist"], "salary.minimum": 60000}`
2. Run `caddie init --answers <file>`, then delete the file.
   - Exit 2: a reply does not fit (the message names the `id`). Ask that
     one question again and rerun.
3. Store each extra answer of your own with
   `caddie answers add --key <key> --scope global "<text>"`.
4. Extra preferences (a board slug, a country) go in `preferences.yml`
   by editing that key only.

## 4. Confirm

1. Run `caddie validate`; it must exit 0.
2. Run `caddie doctor`. Exit 0 is healthy (`warn` is fine); otherwise show
   the user each `fail` line verbatim and stop.
3. If `history/about.md` is missing, tell the user to add their history
   first ([`history/README.md`](../../../../history/README.md)).
