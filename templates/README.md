# templates/

Quarto templates that `caddie render` uses to turn approved Markdown into PDFs
([ADR 0010](../docs/adr/0010-quarto-rendering.md)): Quarto's built-in `typst`
format with Caddie's own partials. No LaTeX, no extension, no network.

| File | Role |
|------|------|
| `resume.qmd` | Resume: `caddie render resume <slug>` → `resumes/<Name>-resume-<position>-<company>.pdf` |
| `cover-letter.qmd` | Letter: `caddie render cover-letter <slug>` → `cover-letters/<Name>-cover-letter-<position>-<company>.pdf` |
| `typst-template.typ` | Shared styling: page, fonts, header, headings, entries, lists, signature |
| `typst-show.typ` | Applies the styling to `_caddie.json` and the qmd's `caddie-layout` |
| `caddie-fields.lua` | Keeps an entry's field lines (organisation, location, dates) on separate lines |

Quarto requires the partial names `typst-template.typ` and `typst-show.typ`.

## Inputs a template receives

`caddie render` copies this directory into a temporary one and writes:

- `_caddie-body.md`: the verified `resumes/<slug>.md` or
  `cover-letters/<slug>.md`, minus any YAML frontmatter. The `.qmd` includes
  it with `{{< include _caddie-body.md >}}`.
- `_caddie.json`: identity from `history/about.md` and contact details from
  `preferences.yml` (**render time only**,
  [ADR 0008](../docs/adr/0008-workspace-tracking.md)). Empty fields are
  left out; `links` is sorted by label:

```json
{"kind": "resume", "name": "Mira Example-Halvorsen", "location": "Utrecht, NL",
 "email": "mira@example.org", "phone": "+00 0 00000000",
 "links": [{"label": "github", "url": "https://github.com/example"}],
 "position": "Senior Data Scientist", "company": "Acme"}
```

The header prints name, location, e-mail, phone and each link's URL without
its scheme. The letter also signs with the name under the closing. `kind`,
`position` and `company` are not printed.

## Previewing by hand

Never render inside `templates/`: Quarto leaves `.quarto/` and output there.

```sh
repo=$PWD tmp=$(mktemp -d) && cp -r templates/. "$tmp" && cd "$tmp"
cp "$repo/testdata/documents/resume-traceable.md" _caddie-body.md
# write a fictional _caddie.json like the one above
quarto render resume.qmd --to typst --output preview.pdf
pdftotext -layout preview.pdf -   # text must be the body plus the header
```

## Contributing

- Templates format; they never add content. Any text that isn't in the
  verified Markdown or `_caddie.json` is a fidelity bug. Separators are
  drawn shapes, not characters; keep it that way.
- Body text stays literal: keep the reader options in the `.qmd` files
  (no smart quotes, citations, math or raw markup) and hyphenation off.
- Use only fonts bundled with Typst (`quarto typst fonts --ignore-system-fonts`).
- Keep each file under 300 lines; shared styling goes in `typst-template.typ`.
- Update the golden text extraction in `testdata/golden/` when the output
  changes on purpose.
- Rendering must work on a headless server with only `quarto` installed.
