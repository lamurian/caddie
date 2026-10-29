package source

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

// keyPath is one configured dotted path and the config key that holds it.
type keyPath struct{ key, path string }

// jsonFieldPaths lists the item fields to read from a JSON listing: the
// portalItem name (id, title…) and its configured key path.
func jsonFieldPaths(f PortalFields) []keyPath {
	return []keyPath{
		{"id", f.ID}, {"title", f.Title}, {"url", f.URL},
		{"location", f.Location}, {"deadline", f.Deadline}, {"salary", f.Salary},
	}
}

// checkJSONPaths rejects malformed dotted paths in a JSON config: empty keys
// (`a..b`) and the HTML-only `@attr` suffix (ADR 0019).
func (p Portal) checkJSONPaths() *problem {
	if p.JSONPath == "" {
		return nil
	}
	paths := []keyPath{{"json_path", p.JSONPath}}
	for _, f := range jsonFieldPaths(p.Fields) {
		paths = append(paths, keyPath{"fields." + f.key, f.path})
	}
	for _, kp := range paths {
		switch {
		case kp.path == "":
		case strings.Contains(kp.path, "@"):
			return &problem{kp.key, "@attr is HTML syntax; use a dotted key path"}
		case slices.Contains(strings.Split(kp.path, "."), ""):
			return &problem{kp.key, fmt.Sprintf("%q has an empty key", kp.path)}
		}
	}
	return nil
}

// extractJSON returns one portalItem per object in the array at p.JSONPath.
// An empty path means the document root is the array. Items without a title
// are dropped; relative URLs resolve against listing_url. A body that is not
// JSON or has no array at the path is an upstream problem wrapping ErrFetch.
func extractJSON(p Portal, body []byte) ([]portalItem, error) {
	fail := func(format string, a ...any) error {
		return fmt.Errorf("%w: %s: json_path %q: %s", ErrFetch, p.SourceName(), p.JSONPath, fmt.Sprintf(format, a...))
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return nil, fail("invalid JSON: %v", err)
	}
	node, err := jsonLookup(doc, p.JSONPath)
	if err != nil {
		return nil, fail("%v", err)
	}
	list, ok := node.([]any)
	if !ok {
		return nil, fail("is %s, not an array", jsonKind(node))
	}
	base, _ := url.Parse(p.ListingURL)
	var items []portalItem
	for i, el := range list {
		obj, ok := el.(map[string]any)
		if !ok {
			return nil, fail("item %d is %s, not an object", i, jsonKind(el))
		}
		if it := jsonItem(obj, p.Fields, base); it["title"] != "" {
			items = append(items, it)
		}
	}
	return items, nil
}

// jsonItem reads each configured field of one listing object; empty values
// are left out of the map.
func jsonItem(obj map[string]any, f PortalFields, base *url.URL) portalItem {
	it := portalItem{}
	for _, kp := range jsonFieldPaths(f) {
		if kp.path == "" {
			continue
		}
		v, err := jsonLookup(obj, kp.path)
		if err != nil {
			continue
		}
		text := jsonText(v)
		if kp.key == "url" && text != "" && base != nil {
			if ref, err := base.Parse(text); err == nil {
				text = ref.String()
			}
		}
		if text != "" {
			it[kp.key] = text
		}
	}
	return it
}

// jsonLookup walks a dotted key path through nested objects; a numeric key
// indexes an array (`column.0`). The empty path is the node itself.
func jsonLookup(node any, path string) (any, error) {
	if path == "" {
		return node, nil
	}
	walked := ""
	for _, key := range strings.Split(path, ".") {
		next, err := jsonStep(node, key, walked)
		if err != nil {
			return nil, err
		}
		node = next
		walked = strings.TrimPrefix(walked+"."+key, ".")
	}
	return node, nil
}

// jsonStep reads key from an object, or a numeric key from an array.
func jsonStep(node any, key, walked string) (any, error) {
	switch n := node.(type) {
	case map[string]any:
		if v, ok := n[key]; ok {
			return v, nil
		}
		return nil, fmt.Errorf("key %q not found in %s", key, jsonWhere(walked))
	case []any:
		i, err := strconv.Atoi(key)
		if err != nil || key[0] == '-' || key[0] == '+' {
			break
		}
		if i >= len(n) {
			return nil, fmt.Errorf("index %d out of range in %s (length %d)", i, jsonWhere(walked), len(n))
		}
		return n[i], nil
	}
	return nil, fmt.Errorf("%s is %s, not an object", jsonWhere(walked), jsonKind(node))
}

func jsonWhere(walked string) string {
	if walked == "" {
		return "the document root"
	}
	return strconv.Quote(walked)
}

// jsonText renders a scalar as trimmed, whitespace-collapsed text; an array
// joins its non-empty scalars with ", ". Null and objects are "".
func jsonText(v any) string {
	switch v := v.(type) {
	case string:
		if arr, ok := jsonArrayString(v); ok {
			return jsonText(arr)
		}
		return strings.Join(strings.Fields(v), " ")
	case json.Number:
		return v.String()
	case bool:
		return strconv.FormatBool(v)
	case []any:
		var parts []string
		for _, el := range v {
			if _, nested := el.([]any); nested {
				continue
			}
			if s := jsonText(el); s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
	}
	return ""
}

// jsonArrayString decodes a string that is itself a JSON array, as Taleo
// sends locations (`["Geneva"]`); any other string is not decoded.
func jsonArrayString(s string) ([]any, bool) {
	t := strings.TrimSpace(s)
	if !strings.HasPrefix(t, "[") || !strings.HasSuffix(t, "]") {
		return nil, false
	}
	dec := json.NewDecoder(strings.NewReader(t))
	dec.UseNumber()
	var arr []any
	if dec.Decode(&arr) != nil || dec.More() {
		return nil, false
	}
	return arr, true
}

// jsonKind names a decoded JSON value's type for error messages.
func jsonKind(v any) string {
	switch v.(type) {
	case map[string]any:
		return "an object"
	case []any:
		return "an array"
	case string:
		return "a string"
	case json.Number:
		return "a number"
	case bool:
		return "a boolean"
	}
	return "null"
}
