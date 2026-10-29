package memory

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// validateRoot checks the shape of answers.yml: a map of global, company
// and posting. Any of them, or the whole file, may be empty (null).
func validateRoot(name string, doc *yaml.Node) error {
	root := doc.Content[0]
	if isNull(root) {
		return nil
	}
	if root.Kind != yaml.MappingNode {
		return malformed(name, root.Line, "the top level must be a map of global, company and posting")
	}
	seen := map[string]bool{}
	for k, v := range pairs(root) {
		if err := checkKey(name, k, seen); err != nil {
			return err
		}
		var err error
		switch k.Value {
		case ScopeGlobal:
			err = validateScope(name, ScopeGlobal, v)
		case ScopeCompany, ScopePosting:
			err = validateNamed(name, k.Value, v)
		default:
			err = malformed(name, k.Line, "unknown top-level key %q (want global, company or posting)", k.Value)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// validateNamed checks company: or posting:, a map of names to scopes.
func validateNamed(name, kind string, v *yaml.Node) error {
	if isNull(v) {
		return nil
	}
	if v.Kind != yaml.MappingNode {
		return malformed(name, v.Line, "%s must be a map of names to answers", kind)
	}
	seen := map[string]bool{}
	for k, m := range pairs(v) {
		if err := checkKey(name, k, seen); err != nil {
			return err
		}
		if err := validateScope(name, kind+":"+k.Value, m); err != nil {
			return err
		}
	}
	return nil
}

// validateScope checks one scope, a map of keys to scalar answers.
func validateScope(name, label string, m *yaml.Node) error {
	if isNull(m) {
		return nil
	}
	if m.Kind != yaml.MappingNode {
		return malformed(name, m.Line, "%s must be a map of keys to answers", label)
	}
	seen := map[string]bool{}
	for k, v := range pairs(m) {
		if err := checkKey(name, k, seen); err != nil {
			return err
		}
		if v.Kind != yaml.ScalarNode {
			return malformed(name, v.Line, "%s: %s must be text, not a list, map or alias", label, k.Value)
		}
	}
	return nil
}

// checkKey rejects a non-text key and a key repeated in the same map,
// which would make the answer ambiguous.
func checkKey(name string, k *yaml.Node, seen map[string]bool) error {
	if k.Kind != yaml.ScalarNode {
		return malformed(name, k.Line, "keys must be plain text")
	}
	if seen[k.Value] {
		return malformed(name, k.Line, "duplicate key %q", k.Value)
	}
	seen[k.Value] = true
	return nil
}

func isNull(n *yaml.Node) bool { return n.Kind == yaml.ScalarNode && n.Tag == "!!null" }

// Set stores text under key in scope s and reports whether it replaced an
// existing value. The node tree is edited in place: a replaced value keeps
// its comments, a new key goes at the end of its scope, and missing scope
// maps are created.
func (a *Answers) Set(s Scope, key, text string) (replaced bool) {
	m := ensureMap(a.ensureRoot(), s.Kind)
	if s.Kind != ScopeGlobal {
		m = ensureMap(m, s.Name)
	}
	if v := lookup(m, key); v != nil {
		setText(v, text)
		return true
	}
	v := &yaml.Node{}
	setText(v, text)
	m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, v)
	return false
}

func (a *Answers) ensureRoot() *yaml.Node {
	if a.doc == nil {
		a.doc = &yaml.Node{Kind: yaml.DocumentNode, HeadComment: a.lead, Content: []*yaml.Node{{}}}
	}
	root := a.doc.Content[0]
	toMap(root)
	return root
}

// ensureMap returns the map under key in parent, creating it, or turning
// an empty (null) value into a map, as needed.
func ensureMap(parent *yaml.Node, key string) *yaml.Node {
	if v := lookup(parent, key); v != nil {
		toMap(v)
		return v
	}
	v := &yaml.Node{}
	toMap(v)
	parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, v)
	return v
}

// toMap turns a validated null node into an empty map, keeping comments.
func toMap(n *yaml.Node) {
	if n.Kind != yaml.MappingNode {
		n.Kind, n.Tag, n.Value, n.Style, n.Content = yaml.MappingNode, "!!map", "", 0, nil
	}
}

// setText makes v a string scalar: double-quoted, or a literal block when
// the text spans lines.
func setText(v *yaml.Node, text string) {
	v.Kind, v.Tag, v.Value, v.Style = yaml.ScalarNode, "!!str", text, yaml.DoubleQuotedStyle
	if strings.Contains(text, "\n") {
		v.Style = yaml.LiteralStyle
	}
}

// Save writes answers.yml atomically (temporary file, then rename) with
// mode 0600, because it may hold salary expectations and referee names.
func (a *Answers) Save(path string) error {
	var buf bytes.Buffer
	if a.doc != nil {
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(a.doc); err != nil {
			return err
		}
		if err := enc.Close(); err != nil {
			return err
		}
	}
	return writeAtomic(path, buf.Bytes())
}

func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(f.Name(), 0o600)
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		_ = os.Remove(f.Name())
	}
	return err
}
