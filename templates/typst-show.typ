// Applies Caddie's styling (typst-template.typ) to the document. Identity
// and contact come from _caddie.json, written by `caddie render`; it is read
// with json() rather than Pandoc metadata, which would parse the values as
// Markdown (an "@" in an e-mail address would become a citation).
#show: caddie-document.with(
  json("_caddie.json"),
$if(caddie-layout)$
  layout: "$caddie-layout$",
$endif$
$if(lang)$
  lang: "$lang$",
$endif$
)
