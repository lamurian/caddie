# 0010. PDF rendering with Quarto + Typst

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §6.2, §6.3, §12.6, D18

## Context

Approved Markdown must become professional PDFs, on a server as well as on a
laptop. LaTeX toolchains are heavy to install.

## Decision

- Render with **Quarto** using a **Typst-based** CV format. The exact format
  is chosen in phase 05 and recorded by amending this ADR.
- Templates live in `templates/resume.qmd` and `templates/cover-letter.qmd`.
  Their paths can be overridden with `CADDIE_RESUME_TEMPLATE` and
  `CADDIE_COVER_LETTER_TEMPLATE`.
- `caddie render` writes the PDF next to its Markdown source and copies it to
  `handoff/`, so the browser container can upload it.
- `quarto` is an external binary invoked through an interface, so tests can
  fake it. Real render tests are skipped when `quarto` is absent.

## Consequences

- Lightweight install: Quarto bundles Typst, and no LaTeX is needed.
- Golden tests compare extracted PDF text, not bytes.
- `caddie doctor` checks that `quarto` is installed and can run.
