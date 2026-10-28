# templates/

Quarto templates that `caddie render` uses to turn approved Markdown into PDFs
([ADR 0010](../docs/adr/0010-quarto-rendering.md)).

| File | Used by | Output |
|------|---------|--------|
| `resume.qmd` | `caddie render resume <slug>` | `resumes/<First>-<Last>-resume-<position>-<company>.pdf` |
| `cover-letter.qmd` | `caddie render cover-letter <slug>` | `cover-letters/<First>-<Last>-cover-letter-<position>-<company>.pdf` |

Both are added in phase [05-render](../docs/tasks/05-render.md). The engine is
Typst through Quarto; no LaTeX.

## Inputs a template receives

- Body: the verified `resumes/<slug>.md` or `cover-letters/<slug>.md`.
- Metadata: identity from `history/about.md` (name, links, languages) and
  contact details (email, phone) merged in from `preferences.yml` **at render
  time only** ([ADR 0008](../docs/adr/0008-workspace-tracking.md)).

## Contributing

- Templates format; they never add content. Any text that isn't in the
  verified Markdown or the metadata is a fidelity bug.
- Keep each file under 300 lines. Put shared styling in a partial
  (e.g. `_style.typ`) rather than duplicating it.
- Update the golden text extraction in `testdata/golden/` when the output
  changes on purpose.
- Rendering must work on a headless server with only `quarto` installed.
