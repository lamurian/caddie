package source

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Kind is the role of a text block in a job description.
type Kind int

// Block kinds.
const (
	Para Kind = iota
	Heading
	Item
)

// Block is one paragraph, heading or list item of plain text.
type Block struct {
	Kind Kind
	Text string
}

// maxHeading is the longest bold-only or colon-ended line taken as a heading.
const maxHeading = 80

// Blocks turns description HTML into plain-text blocks. Headings are h1–h6
// elements, paragraphs whose text is all bold, and short paragraphs ending
// in a colon (ADR 0017). Scripts and styles are dropped. Text that was
// entity-escaped HTML (as Greenhouse returns it) is unescaped first.
func Blocks(s string) []Block {
	if !strings.Contains(s, "<") && strings.Contains(s, "&lt;") {
		s = html.UnescapeString(s)
	}
	root, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return nil
	}
	w := &walker{}
	w.children(root)
	w.flush()
	return w.blocks
}

// Text is the plain text of description HTML, blocks joined by spaces.
func Text(s string) string {
	var parts []string
	for _, b := range Blocks(s) {
		parts = append(parts, b.Text)
	}
	return strings.Join(parts, " ")
}

type walker struct {
	blocks      []Block
	buf         strings.Builder
	kind        Kind
	bold        int // depth of <b>/<strong>
	boldRunes   int
	plainRunes  int
	headingTags int
}

var skipped = map[atom.Atom]bool{atom.Script: true, atom.Style: true, atom.Noscript: true,
	atom.Head: true, atom.Template: true, atom.Svg: true, atom.Button: true, atom.Form: true,
	atom.Nav: true}

var headings = map[atom.Atom]bool{atom.H1: true, atom.H2: true, atom.H3: true,
	atom.H4: true, atom.H5: true, atom.H6: true}

var blockTags = map[atom.Atom]bool{atom.P: true, atom.Div: true, atom.Section: true,
	atom.Article: true, atom.Ul: true, atom.Ol: true, atom.Table: true, atom.Tr: true,
	atom.Blockquote: true, atom.Main: true, atom.Header: true, atom.Footer: true,
	atom.Dd: true, atom.Dt: true, atom.Pre: true, atom.Li: true, atom.Body: true}

func (w *walker) children(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		w.node(c)
	}
}

func (w *walker) node(n *html.Node) {
	switch n.Type {
	case html.TextNode:
		w.text(n.Data)
	case html.ElementNode:
		w.element(n)
	case html.DocumentNode:
		w.children(n)
	}
}

func (w *walker) text(s string) {
	w.buf.WriteString(s)
	n := utf8.RuneCountInString(strings.Join(strings.Fields(s), ""))
	if w.bold > 0 {
		w.boldRunes += n
	} else {
		w.plainRunes += n
	}
}

func (w *walker) element(n *html.Node) {
	tag := n.DataAtom
	switch {
	case skipped[tag]:
		return
	case tag == atom.Br:
		w.flush()
	case tag == atom.B || tag == atom.Strong:
		w.bold++
		w.children(n)
		w.bold--
	case headings[tag]:
		w.flush()
		w.headingTags++
		w.children(n)
		w.flush()
		w.headingTags--
	case blockTags[tag]:
		w.blockElement(n, tag)
	default:
		w.children(n)
	}
}

func (w *walker) blockElement(n *html.Node, tag atom.Atom) {
	w.flush()
	prev := w.kind
	if tag == atom.Li {
		w.kind = Item
	}
	w.children(n)
	w.flush()
	w.kind = prev
}

func (w *walker) flush() {
	text := strings.Join(strings.Fields(w.buf.String()), " ")
	allBold := w.boldRunes > 0 && w.plainRunes == 0
	w.buf.Reset()
	w.boldRunes, w.plainRunes = 0, 0
	if text == "" {
		return
	}
	kind := w.kind
	short := utf8.RuneCountInString(text) <= maxHeading
	if w.headingTags > 0 || (kind == Para && short && (allBold || strings.HasSuffix(text, ":"))) {
		kind = Heading
		text = strings.TrimSpace(strings.TrimSuffix(text, ":"))
	}
	w.blocks = append(w.blocks, Block{Kind: kind, Text: text})
}
