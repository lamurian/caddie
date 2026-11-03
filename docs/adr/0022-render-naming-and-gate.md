# 0022. Render: PDF naming, the verified gate and the render directory

- **Status:** Accepted
- **Date:** 2026-10-29
- **PRD refs:** §6.2.6, §6.3.6, §12.6, D9, D15, D18

## Context

ADR 0010 picks Quarto + Typst and ADR 0004 makes rendering depend on a
verified hash. Phase 05 must fix what the PDF is called, where the gate is
enforced, how identity and contact reach the template without leaking into
tracked files, how the posting's status moves, and what path a verified
record stores. PRD §12.6 says `<Firstname>-<Lastname>-…` but names do not
split cleanly into first and last.

## Decision

- **Naming.** `PDFName` returns
  `<Name>-resume-<position-slug>-<company-slug>.pdf` or
  `<Name>-cover-letter-…`. `<Name>` is `about.md` `name` with diacritics
  folded, case kept, every run of characters other than ASCII letters and
  digits replaced by one hyphen, and no leading or trailing hyphen
  ("José Ñúñez O'Brien" → `Jose-Nunez-O-Brien`). It is capped at 64
  characters on a hyphen so the whole name stays under 255 bytes. Position
  and company use `workspace.Slug` (lowercase, capped at 80). An empty part
  is left out instead of leaving a stray hyphen.
- **Gate in `Render`.** `document.Render` itself reads the draft and
  refuses with `ErrNotVerified` unless `IsVerified` matches its exact
  bytes, before it makes a temp dir or calls quarto. The CLI maps this to
  exit 7 (refused by a safety rule) and the message names the draft and the
  fix, `run caddie verify <path>`. A failed or output-less quarto run wraps
  `ErrQuarto` with the last lines of its output: exit 8.
- **Render directory.** `Render` copies every regular file of the
  template's directory (recursively, dot-entries skipped) into
  `os.MkdirTemp("", "caddie-render-")`, writes two files there and runs
  `quarto render <template base> --to typst --output <PDFName>`; the
  directory is always removed afterwards.
  - `_caddie-body.md`: the verified draft minus a leading `---` frontmatter
    block, included by the `.qmd`.
  - `_caddie.json`: `kind`, `name`, `location`, `email`, `phone`, `links`
    (sorted by label), `position`, `company`; empty fields are omitted. The
    Typst partial reads it with `json("_caddie.json")`. It is JSON, not
    pandoc metadata, because pandoc parses metadata as Markdown and turns
    the `@` of an e-mail address into a citation.
- **Literal body.** Quarto would expand `{{< env/include/meta >}}`, read
  mid-body YAML blocks and run code cells, so `_caddie-body.md` gets
  `{{<` → `{\{<`, `<!--` → `<\!--` and `---` lines → `----`, and the
  templates set `engine: markdown`: only verified text reaches the PDF.
- **Contact** (e-mail, phone from `preferences.yml`) is written only into
  that temp dir and so only reaches the PDF, never a tracked file (ADR 0008).
- **Handoff.** The PDF goes next to the draft and is copied to `handoff/`
  (created if missing). Both overwrite an earlier render of the same name.
- **Status.** `caddie render` advances `postings/<slug>.md`: `saved` →
  `tailored` → `rendered` (both steps), `tailored` → `rendered`; `rendered`
  stays `rendered`, so the resume and the letter can each be rendered.
  `applied`, `manual` or a posting outside `postings/` is refused with
  `workspace.ErrIllegalTransition` (exit 6) before quarto runs.
- **Verified records.** `.caddie/verified/<sha256>` stores `path` as the
  draft's path relative to the workspace root, slash-separated
  (`resumes/<slug>.md`), or the cleaned absolute path when the draft lies
  outside the workspace. Only the hash is authoritative; `path` is for
  people reading the record.

## Consequences

- Any caller of `Render`, not only the CLI, gets the gate: it cannot be
  skipped by calling the package directly (ADR 0004 stays intact).
- Tests drive `Render` with a fake `Quarto`; only the real-render golden test
  needs the binary and is skipped without it.
- Names in scripts without ASCII letters (e.g. CJK) fold to nothing, and
  letters with no decomposition (`ø`, `ß`, `Ł`) become hyphens; the file is
  then named without the `<Name>` part or with a split name. Accepted for
  now; the PDF content is unaffected.
- Templates must read identity and contact from `_caddie.json` and the body
  from `_caddie-body.md`; renaming either is a breaking change for
  templates overridden with `CADDIE_*_TEMPLATE`.
- Rendering the same draft twice overwrites both copies, which is intended.
