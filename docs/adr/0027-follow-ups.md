# 0027. Follow-ups: boundaries, past deadlines and report shape

- **Status:** Accepted
- **Date:** 2026-11-03
- **PRD refs:** §6.5, §12.3, §12.7

## Context

PRD §6.5 says `caddie follow-ups` reports applied entries "past"
`follow_up_after`, manual entries with open checkboxes, and postings whose
`deadline` is "within" `apply.deadline_warning_days`. It does not fix the
boundary days, what happens to a deadline already gone, or the output.

## Decision

- **Library.** `workspace.FollowUps(layout, warnDays, today)` builds a
  `FollowUpReport`; the clock is injected (never `time.Now` inside). Day
  counts use the calendar day of `today` in its own location; frontmatter
  dates are plain `YYYY-MM-DD` calendar days.
- **Which files.** Every `<dir>/*.md` directly in `applied/`, `manual/` and
  `postings/`, except `README.md`, dot files and the `<slug>/` artefact
  directories (ADR 0026). Directories decide the category, not `status`.
  A missing directory counts as empty.
- **Follow-ups** (`applied/` only): due when `today >= follow_up_after`;
  `days_overdue` is 0 on the day itself. Unset `follow_up_after` → skipped.
- **Manual steps** (`manual/` only): the unticked `- [ ]` / `* [ ]` items
  of the `## Manual steps` section that `caddie posting steps` writes,
  ending at the next `#`/`##` heading. Items elsewhere in the body are not
  steps. Reported with `open` (the count) and `items` (their text, in file
  order) when at least one is open.
- **Deadlines** (`postings/` and `manual/`, never `applied/`):
  `days_left = deadline − today`. Reported when
  `days_left <= deadline_warning_days`, so the boundary day counts and a
  deadline today is `days_left: 0`, still open. A deadline already gone
  (`days_left < 0`) is always reported, whatever the window, with
  `past: true`. Unset `deadline` → skipped.
- **Errors.** Malformed frontmatter, or a date that is not `YYYY-MM-DD`,
  wraps `ErrBadPosting` naming the workspace-relative file and the field
  (exit 4). A negative `warnDays` wraps `config.ErrConfig` (exit 3).
- **JSON.** `{"today", "deadline_warning_days", "follow_ups": [...],
  "manual_steps": [...], "deadlines": [...]}`. Each entry carries `slug`,
  `title`, `company` and `posting` (workspace-relative path), plus
  `applied_at`, `follow_up_after`, `days_overdue`; or `manual_reason`,
  `open`, `items`; or `status`, `deadline`, `days_left`, `past`. Lists are
  never `null`.
- **Order.** Follow-ups by `follow_up_after`, deadlines by `deadline`
  (oldest, i.e. most urgent, first), manual steps by path; ties by path.

## Consequences

- A posting can appear twice (a `manual/` entry with open steps and a near
  deadline); each list answers one question.
- Past deadlines in `postings/` keep showing until the user removes or
  moves the posting, which is the prompt to do so.
- The checklist reader in `internal/workspace` repeats the section rule of
  `internal/posting` (which imports `workspace`, so it cannot be reused);
  a change to the `## Manual steps` format must update both.
