# 0020. Answers: scopes, lookup, form keys and round-trip editing

- **Status:** Accepted
- **Date:** 2026-10-29
- **PRD refs:** §6.3, §6.4, §7, §12.5, D13

## Context

ADR 0007 fixes the three scopes of `answers.yml` and the lookup order
posting → company → global. It leaves open how a scope is written on the
command line, how a posting finds its company, which answers belong on an
application form, what `caddie answers` prints, and how a file the user
edits by hand survives Caddie writing to it.

## Decision

- **Scope syntax:** `--scope global`, `company:<name>` or `posting:<slug>`.
  In the file they are the maps `global`, `company.<name>` and
  `posting.<slug>`. A company name is stored as its slug
  (`company:Société Générale` → `societe-generale`); a posting scope must
  already be a slug. Keys match `[a-z0-9_]+`; the text is trimmed and must
  not be empty.
- **Company key** = `workspace.Slug` of the posting's `company`
  frontmatter. A hand-written company key must be that slug to be found.
- **`--for <slug>`** reads the posting from `postings/<slug>.md`, else
  `manual/<slug>.md`, else `applied/<slug>.md`; none is exit 5. The merged
  view holds one answer per key, posting beating company beating global,
  sorted by key, each with the scope it came from. A key whose value is
  empty counts as unanswered and does not hide a broader answer.
- **`--form`** leaves out the cover-letter-only interview keys `why_org`,
  `why_role` and `lead_story`; everything else (including `start_date` and
  `constraints`) may be used on forms.
- **Output** is always JSON (`--json` is accepted and ignored):
  - `add`: `{"scope","key","value","replaced"}`.
  - `show`: `{"for","company","answers":[{"key","value","scope"}]}`;
    without `--for` only global answers are shown and `for`/`company` are
    left out.
  - `list`: `{"answers":[…]}`, every answer in every scope, empty ones
    included, in file order.
- **Hand-edit tolerance:** a value may be any scalar and keeps its literal
  text (`85000`, `yes`); a null is empty. Anything else is malformed, exit
  4 naming `file:line`: a top level that is not a map, an unknown top-level
  key, a scope that is not a map, a list, map or alias as a value, a
  repeated key, or a second YAML document.
- **Round-trip:** `answers add` edits the parsed `go.yaml.in/yaml/v3` node
  tree in place and re-encodes it. Key order, quoting style and head,
  line and foot comments survive; a replaced value keeps its comments. A
  new key goes at the end of its scope; missing scope maps are created. A
  new value is double-quoted, or a `|` block when it spans lines. The file
  is written atomically (temporary file, rename) with mode 0600.

## Consequences

- Limitations of re-encoding with yaml.v3: blank lines are dropped,
  indentation becomes two spaces, and flow style or unusual comment
  placement may be normalised. Comments and keys are never lost; a file
  holding only comments keeps them as a header.
- The agent gets the same answer for the same posting from any directory
  the posting has moved to, and never sees an interview story on a form.
- A company stored under a non-slug key by hand is invisible to `--for`
  until renamed; `answers list` shows it, so the user can spot it.
