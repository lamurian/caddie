// Caddie's shared Typst styling for the resume and the cover letter
// (ADR 0010). It formats only: every word on the page comes from the body
// (_caddie-body.md) or from _caddie.json (name, location, email, phone,
// links). Fonts are the ones bundled with Typst; nothing is downloaded.

#let caddie-accent = rgb("#1f3a5f")
#let caddie-muted = luma(80)
#let caddie-ink = luma(15)

// Field value, or none when missing or empty.
#let caddie-get(info, key) = {
  let v = info.at(key, default: none)
  if v == "" { none } else { v }
}

// A drawn dot between header fields: a shape, not a character, so text
// extraction shows the fields and nothing else.
#let caddie-sep = {
  h(0.6em)
  box(baseline: -0.3em, circle(radius: 0.1em, fill: caddie-muted))
  h(0.6em)
}

// "https://github.com/example/" -> "github.com/example".
#let caddie-url-text(url) = {
  url.replace(regex("^[A-Za-z][A-Za-z0-9+.-]*://"), "").trim("/", at: end)
}

#let caddie-contact(info) = {
  let fields = ()
  let location = caddie-get(info, "location")
  if location != none { fields.push(location) }
  let email = caddie-get(info, "email")
  if email != none { fields.push(link("mailto:" + email, email)) }
  let phone = caddie-get(info, "phone")
  if phone != none { fields.push(phone) }
  for l in info.at("links", default: ()) {
    let url = l.at("url", default: "")
    if url != "" { fields.push(link(url, caddie-url-text(url))) }
  }
  fields
}

#let caddie-header(info, rule: false) = {
  let name = caddie-get(info, "name")
  let fields = caddie-contact(info)
  block(
    width: 100%,
    below: if rule { 2.4em } else { 1.2em },
    inset: (bottom: if rule { 0.8em } else { 0pt }),
    stroke: if rule { (bottom: 0.6pt + caddie-accent) } else { none },
  )[
    #if name != none {
      set par(leading: 0.4em)
      text(size: 2.2em, weight: "semibold", fill: caddie-accent, name)
      v(0.15em)
    }
    #if fields.len() > 0 {
      set par(leading: 0.45em)
      text(size: 0.9em, fill: caddie-muted, fields.join(caddie-sep))
    }
  ]
}

// The field lines under an entry title, wrapped by caddie-fields.lua.
#let caddie-fields(body) = block(above: 0.3em, below: 0.45em, sticky: true)[
  #set par(leading: 0.42em)
  #set text(size: 0.93em, fill: caddie-muted, style: "italic")
  #body
]

#let caddie-document(info, layout: "resume", lang: "en", doc) = {
  let letter = layout == "letter"
  let name = caddie-get(info, "name")
  set document(title: name, author: if name == none { () } else { name })
  set page(
    paper: "a4",
    margin: (x: if letter { 2.6cm } else { 2.2cm }, top: 2cm, bottom: 2cm),
    numbering: none,
    header: none,
    footer: none,
  )
  set text(
    font: "Libertinus Serif",
    size: if letter { 11pt } else { 10.5pt },
    lang: lang,
    hyphenate: false,
    fill: caddie-ink,
  )
  set smartquote(enabled: false)
  set par(
    justify: false,
    leading: if letter { 0.72em } else { 0.55em },
    spacing: if letter { 1.35em } else { 0.75em },
  )
  set list(indent: 0.25em, body-indent: 0.55em, spacing: 0.42em,
    marker: text(fill: caddie-accent)[•])
  set enum(indent: 0.25em, body-indent: 0.55em, spacing: 0.42em)
  show link: set text(fill: caddie-accent)
  show heading: set text(fill: caddie-accent)
  show heading.where(level: 1): it => block(above: 1.4em, below: 0.8em,
    sticky: true, text(size: 1.3em, weight: "bold", it.body))
  show heading.where(level: 2): it => block(
    width: 100%,
    above: 1.25em,
    below: 0.6em,
    inset: (bottom: 0.25em),
    stroke: (bottom: 0.6pt + caddie-accent),
    sticky: true,
    text(size: 1.15em, weight: "bold", it.body),
  )
  show heading.where(level: 3): it => block(above: 0.95em, below: 0.3em,
    sticky: true, text(size: 1.02em, weight: "bold", fill: caddie-ink, it.body))
  show heading.where(level: 4): it => block(above: 0.8em, below: 0.3em,
    sticky: true, text(weight: "bold", style: "italic", it.body))

  caddie-header(info, rule: letter)
  doc
  // The letter body ends at the closing ("Kind regards,"); sign under it.
  if letter and name != none {
    block(above: 2.8em, breakable: false, name)
  }
}
