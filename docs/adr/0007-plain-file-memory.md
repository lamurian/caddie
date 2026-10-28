# 0007. Plain-file, user-editable memory

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §6.1, §6.3, §12.5, D13

## Context

The agent must not re-rank jobs the user has already seen or rejected, and
must not ask the same interview question twice. Hidden state such as a
database or an embedding store is opaque to the user.

## Decision

- `.caddie/seen.jsonl` stores one JSON object per line: `id`, `url`, `title`,
  `company`, `status` (`shown|rejected|saved`) and `at`. `caddie search`
  excludes seen IDs unless `--include-seen` is passed.
- `answers.yml` has three scopes: `global`, `company.<x>` and
  `posting.<slug>`. Lookup order is posting → company → global.
- Both are plain files, git-ignored, and readable and editable by the user.
  They are written only through `caddie seen` and `caddie answers`.

## Consequences

- The user can inspect, correct or delete memory with a text editor, which
  makes it more trustworthy than hidden state.
- The files must tolerate hand edits: parse leniently, report clearly, and
  never silently drop lines.
- `answers.yml` may contain salary expectations and referee names, so it is
  never committed. `answers.example.yml` documents the shape.
