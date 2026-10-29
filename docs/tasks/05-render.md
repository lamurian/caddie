# 05. Render

- **Status:** Done
- **PRD milestone:** M3 (part 2)

## Goal

An approved and verified resume or cover letter becomes a well-typeset PDF
under the fixed naming scheme, with a copy in `handoff/` ready for upload.

## Inputs

- PRD §6.2.6, §6.3.6, §12.6, §13
- ADRs [0010](../adr/0010-quarto-rendering.md),
  [0004](../adr/0004-fidelity-check.md),
  [0008](../adr/0008-workspace-tracking.md)

## Scope

- Choose the Quarto Typst CV format, then amend ADR 0010 with the choice.
- Add `templates/resume.qmd` and `templates/cover-letter.qmd`. Shared styling
  goes in a partial.
- `internal/document/render` runs the Quarto binary behind an interface. It
  merges identity from `about.md` and contact details from `preferences.yml`
  at render time only.
- `caddie render resume|cover-letter <slug>`:
  - refuses unless the verified hash matches the current content;
  - writes `<First>-<Last>-resume-<position>-<company>.pdf` (or
    `-cover-letter-`) next to the source;
  - copies the PDF to `handoff/`;
  - advances the posting status to `rendered`.
- `caddie doctor` checks that `quarto` exists and can render a trivial
  document.

## Out of scope

Uploading the PDF (phase 07).

## Deliverables

`templates/*.qmd` (plus partials), `internal/document/render*.go`, a golden
text extraction in `testdata/golden/` and skill updates to the render steps.

## Definition of Done

- [x] `render` on unverified or edited content exits non-zero with a clear
      message (tested with a fake Quarto).
- [x] File naming is tested, including names and companies with spaces,
      accents and punctuation.
- [x] Contact details appear in the PDF but never in any tracked file or in
      `caddie profile` output.
- [x] The real render test is skipped without `quarto`. When it runs, the
      extracted text matches the golden file.
- [x] The `handoff/` copy exists after rendering; the posting status becomes
      `rendered`.
- [x] Rendering works on a headless Linux box with only Quarto installed
      (noted in the PR; verified in a headless container with Quarto
      1.10.19 from mise, no LaTeX, and with the network cut off).
- [x] The resume draft format (name, headline, contact) is settled with
      `verify`: `#`/`##` heading text is not checked (ADR 0021), so the
      template supplies name and contact and drafts carry none; update
      `references/tailor.md` and the `testdata/documents/` fixtures to match.
- [x] One path convention for `.caddie/verified/` records (verify stores the
      path as given, relative to the current directory).
