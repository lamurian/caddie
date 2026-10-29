package source

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"

	"github.com/lamurian/caddie/internal/config"
)

// portalItem is one listing entry before normalisation: field name
// (id, title, url, location, deadline, salary) to its whitespace-collapsed text.
type portalItem map[string]string

// fieldSel is one compiled field of an HTML portal: sel finds the element
// inside the item (nil means the item itself) and attr, when set, is read
// instead of the element's text.
type fieldSel struct {
	name string
	sel  cascadia.Selector
	attr string
}

// htmlSels is a portal's compiled item selector and field selectors.
type htmlSels struct {
	item   cascadia.Selector
	fields []fieldSel
}

// extractHTML returns one portalItem per element matching p.ItemSelector.
// Items appear in document order; one with an empty title is dropped. A
// field selector without a match gives an empty value, and the url field is
// resolved against p.ListingURL (ADR 0019).
func extractHTML(p Portal, body []byte) ([]portalItem, error) {
	sels, pr := compileHTML(p)
	if pr != nil {
		return nil, fmt.Errorf("%w: %s: %s: %s", config.ErrConfig, p.File, pr.field, pr.msg)
	}
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrFetch, p.SourceName(), err)
	}
	base, _ := url.Parse(p.ListingURL)
	var items []portalItem
	for _, n := range sels.item.MatchAll(root) {
		it := portalItem{}
		for _, f := range sels.fields {
			it[f.name] = f.value(n)
		}
		if it["title"] == "" {
			continue
		}
		if raw := it["url"]; raw != "" && base != nil {
			it["url"] = resolveURL(base, raw)
		}
		items = append(items, it)
	}
	return items, nil
}

// checkSelectors rejects an HTML config whose selectors do not compile.
func (p Portal) checkSelectors() *problem {
	if p.ItemSelector == "" {
		return nil
	}
	_, pr := compileHTML(p)
	return pr
}

// compileHTML compiles the item selector and every non-empty field.
func compileHTML(p Portal) (htmlSels, *problem) {
	item, err := cascadia.Compile(p.ItemSelector)
	if err != nil {
		return htmlSels{}, &problem{"item_selector", err.Error()}
	}
	sels := htmlSels{item: item}
	f := p.Fields
	for _, kv := range [][2]string{{"id", f.ID}, {"title", f.Title}, {"url", f.URL},
		{"location", f.Location}, {"deadline", f.Deadline}, {"salary", f.Salary}} {
		if kv[1] == "" {
			continue
		}
		fs, err := compileField(kv[0], kv[1])
		if err != nil {
			return htmlSels{}, &problem{"fields." + kv[0], err.Error()}
		}
		sels.fields = append(sels.fields, fs)
	}
	return sels, nil
}

// compileField splits "selector@attr" and compiles the selector part. The
// last "@" starts an attribute only if what follows is a bare name, so an
// "@" inside an attribute-value selector (a[href*="@"]) stays in the selector.
func compileField(name, spec string) (fieldSel, error) {
	fs := fieldSel{name: name}
	sel := spec
	if i := strings.LastIndex(spec, "@"); i >= 0 && isAttrName(spec[i+1:]) {
		sel, fs.attr = strings.TrimSpace(spec[:i]), strings.ToLower(spec[i+1:])
	}
	if sel == "" {
		return fs, nil
	}
	compiled, err := cascadia.Compile(sel)
	if err != nil {
		return fs, err
	}
	fs.sel = compiled
	return fs, nil
}

func isAttrName(s string) bool {
	return s != "" && !strings.ContainsAny(s, " \t\n]\"'=()[,>+~:.#*")
}

// value is the field's whitespace-collapsed text or attribute in item; the
// item itself counts as a match for its own selector.
func (f fieldSel) value(item *html.Node) string {
	n := item
	if f.sel != nil {
		if n = f.sel.MatchFirst(item); n == nil {
			return ""
		}
	}
	s := attr(n, f.attr)
	if f.attr == "" {
		s = nodeText(n)
	}
	return strings.Join(strings.Fields(s), " ")
}

// resolveURL resolves a link from the listing against its URL; a link that
// does not parse is kept as written.
func resolveURL(base *url.URL, raw string) string {
	ref, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return base.ResolveReference(ref).String()
}
