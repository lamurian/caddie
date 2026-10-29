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

## Amendment (2026-10-29): the format

- **Quarto's built-in `typst` format with Caddie's own partials**, not a
  Quarto CV extension. `templates/typst-template.typ` holds the styling
  both documents share; `templates/typst-show.typ` applies it.
  `resume.qmd` and `cover-letter.qmd` differ only in `caddie-layout`.
- Identity and contact reach Typst as `_caddie.json`, read with `json()`;
  the verified body arrives as `_caddie-body.md` through
  `{{< include >}}`. Both are written into a temporary render directory.
  Pandoc metadata is not used: it parses values as Markdown, so an `@` in
  an e-mail address would become a citation.
- The reader keeps the body literal (`-smart-citations-tex_math_dollars`,
  no raw TeX/HTML/attributes, Typst smart quotes and hyphenation off) and
  accepts a list right after a line (`+lists_without_preceding_blankline`),
  as history entries are written. `caddie-fields.lua` keeps an entry's
  field lines (organisation, location, dates) on separate lines.
- Fonts: Libertinus Serif, bundled with Typst. Header separators are drawn
  shapes, not characters. No page numbers.
- Why not an extension such as `quarto-awesomecv-typst` or `modern-cv`:
  each must be installed with `quarto add` (network, and vendored code to
  keep updated), needs Font Awesome or other fonts not bundled with Typst,
  and adds icons and fixed text (section labels, dates) that would
  bypass the fidelity check.

### Consequences

- Rendering works offline on a headless box with only `quarto` installed;
  a render takes about 1.5 s.
- The extracted text of a PDF is exactly the body plus name, location,
  e-mail, phone and links (and, for a letter, the name under the closing),
  which keeps golden tests and contact-leak tests simple.
- The look is ours to maintain: changes go in `typst-template.typ`.
