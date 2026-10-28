package source

import (
	"bytes"
	"encoding/json"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// page holds what pageDetail reads from an HTML document.
type page struct {
	jsonLD          []string
	meta            map[string]string
	title, h1       string
	main, art, body *html.Node
}

// pageDetail reads a posting page. schema.org JobPosting JSON-LD wins;
// otherwise the title comes from og:title, <h1> or <title> and the
// description from <main>, <article> or <body>.
func pageDetail(doc string) Detail {
	root, err := html.Parse(strings.NewReader(doc))
	if err != nil {
		return Detail{}
	}
	p := &page{meta: map[string]string{}}
	p.scan(root)
	for _, raw := range p.jsonLD {
		if d, ok := jobPosting(raw); ok {
			return d
		}
	}
	d := Detail{Job: Job{
		Title:   firstNonEmpty(p.meta["og:title"], p.h1, p.title),
		Company: p.meta["og:site_name"],
	}}
	for _, n := range []*html.Node{p.main, p.art, p.body} {
		if n != nil {
			var b bytes.Buffer
			_ = html.Render(&b, n)
			d.HTML = b.String()
			break
		}
	}
	return d
}

func (p *page) scan(n *html.Node) {
	if n.Type == html.ElementNode {
		p.element(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.scan(c)
	}
}

func (p *page) element(n *html.Node) {
	switch n.DataAtom {
	case atom.Script:
		if attr(n, "type") == "application/ld+json" && n.FirstChild != nil {
			p.jsonLD = append(p.jsonLD, n.FirstChild.Data)
		}
	case atom.Meta:
		if k := firstNonEmpty(attr(n, "property"), attr(n, "name")); k != "" {
			p.meta[k] = strings.TrimSpace(attr(n, "content"))
		}
	case atom.Title:
		p.title = strings.TrimSpace(nodeText(n))
	case atom.H1:
		if p.h1 == "" {
			p.h1 = strings.Join(strings.Fields(nodeText(n)), " ")
		}
	case atom.Main:
		p.main = n
	case atom.Article:
		if p.art == nil {
			p.art = n
		}
	case atom.Body:
		p.body = n
	}
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func nodeText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}

// jobPosting finds a schema.org JobPosting in one JSON-LD script, which may
// hold an object, an array or an @graph.
func jobPosting(raw string) (Detail, bool) {
	var v any
	if json.Unmarshal([]byte(raw), &v) != nil {
		return Detail{}, false
	}
	for _, obj := range ldObjects(v) {
		if isType(obj["@type"], "JobPosting") {
			return ldDetail(obj), true
		}
	}
	return Detail{}, false
}

func ldObjects(v any) []map[string]any {
	switch t := v.(type) {
	case []any:
		var out []map[string]any
		for _, x := range t {
			out = append(out, ldObjects(x)...)
		}
		return out
	case map[string]any:
		return append([]map[string]any{t}, ldObjects(t["@graph"])...)
	}
	return nil
}

func isType(v any, want string) bool {
	switch t := v.(type) {
	case string:
		return t == want
	case []any:
		for _, x := range t {
			if s, _ := x.(string); s == want {
				return true
			}
		}
	}
	return false
}
