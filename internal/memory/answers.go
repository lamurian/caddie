package memory

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

// ErrMalformedAnswers marks an answers.yml whose structure Caddie cannot
// use; the message names file and line.
var ErrMalformedAnswers = errors.New("malformed answers file")

// Scope kinds: the three top-level maps of answers.yml (ADR 0007).
const (
	ScopeGlobal  = "global"
	ScopeCompany = "company"
	ScopePosting = "posting"
)

// Scope says where an answer applies: global, company:<x> or
// posting:<slug>. Name is empty for global.
type Scope struct {
	Kind string
	Name string
}

func (s Scope) String() string {
	if s.Kind == ScopeGlobal {
		return ScopeGlobal
	}
	return s.Kind + ":" + s.Name
}

// ParseScope reads `global`, `company:<x>` or `posting:<slug>`.
func ParseScope(s string) (Scope, error) {
	kind, name, hasName := strings.Cut(s, ":")
	switch {
	case kind == ScopeGlobal && !hasName:
		return Scope{Kind: ScopeGlobal}, nil
	case (kind == ScopeCompany || kind == ScopePosting) && strings.TrimSpace(name) != "":
		return Scope{Kind: kind, Name: strings.TrimSpace(name)}, nil
	}
	return Scope{}, fmt.Errorf("scope %q: want global, company:<name> or posting:<slug>", s)
}

// Answer is one stored answer and the scope it came from.
type Answer struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Scope string `json:"scope"`
}

// CoverLetterKeys are interview answers used only in cover letters;
// FormAnswers leaves them out (ADR 0020).
var CoverLetterKeys = map[string]bool{"why_org": true, "why_role": true, "lead_story": true}

// FormAnswers returns the answers usable on application forms.
func FormAnswers(list []Answer) []Answer {
	out := []Answer{}
	for _, a := range list {
		if !CoverLetterKeys[a.Key] {
			out = append(out, a)
		}
	}
	return out
}

// Answers is answers.yml as a YAML node tree, so that saving keeps the
// user's key order and comments.
type Answers struct {
	doc  *yaml.Node // the document node; nil while the file is empty
	lead string     // the comments of a file that holds nothing else
}

// LoadAnswers reads answers.yml. A missing or empty file has no answers.
func LoadAnswers(path string) (*Answers, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Answers{}, nil
	}
	if err != nil {
		return nil, err
	}
	return ParseAnswers(path, data)
}

// ParseAnswers parses answers.yml; name is used in error messages. Values
// may be any scalar and keep their literal text; anything else in the
// wrong place is an ErrMalformedAnswers naming file and line.
func ParseAnswers(name string, data []byte) (*Answers, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var doc yaml.Node
	if err := dec.Decode(&doc); errors.Is(err, io.EOF) {
		return &Answers{lead: strings.TrimSpace(string(data))}, nil
	} else if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrMalformedAnswers, name, err)
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, malformed(name, extra.Line, "only one YAML document is allowed")
	}
	if err := validateRoot(name, &doc); err != nil {
		return nil, err
	}
	return &Answers{doc: &doc}, nil
}

func malformed(name string, line int, format string, args ...any) error {
	return fmt.Errorf("%w: %s:%d: %s", ErrMalformedAnswers, name, line, fmt.Sprintf(format, args...))
}

// For is the merged view for a posting: one answer per key, posting:<slug>
// beating company:<company> beating global, sorted by key. An empty slug or
// company skips that scope; an empty value counts as unanswered.
func (a *Answers) For(slug, company string) []Answer {
	scopes := []Scope{{Kind: ScopeGlobal}}
	if company != "" {
		scopes = append(scopes, Scope{Kind: ScopeCompany, Name: company})
	}
	if slug != "" {
		scopes = append(scopes, Scope{Kind: ScopePosting, Name: slug})
	}
	byKey := map[string]Answer{}
	for _, s := range scopes {
		for _, ans := range a.scopeAnswers(s, a.scopeMap(s)) {
			if ans.Value != "" {
				byKey[ans.Key] = ans
			}
		}
	}
	out := make([]Answer, 0, len(byKey))
	for _, ans := range byKey {
		out = append(out, ans)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// All lists every stored answer in every scope, in file order, including
// empty ones.
func (a *Answers) All() []Answer {
	out := []Answer{}
	for kind, section := range pairs(a.root()) {
		if kind.Value == ScopeGlobal {
			out = append(out, a.scopeAnswers(Scope{Kind: ScopeGlobal}, section)...)
			continue
		}
		for name, m := range pairs(section) {
			out = append(out, a.scopeAnswers(Scope{Kind: kind.Value, Name: name.Value}, m)...)
		}
	}
	return out
}

func (a *Answers) scopeAnswers(s Scope, m *yaml.Node) []Answer {
	var out []Answer
	for k, v := range pairs(m) {
		out = append(out, Answer{Key: k.Value, Value: scalarText(v), Scope: s.String()})
	}
	return out
}

// scalarText is a value's literal text; a YAML null is empty.
func scalarText(v *yaml.Node) string {
	if v.Tag == "!!null" {
		return ""
	}
	return v.Value
}

// root is the top-level mapping, or nil.
func (a *Answers) root() *yaml.Node {
	if a.doc == nil || len(a.doc.Content) == 0 {
		return nil
	}
	return a.doc.Content[0]
}

// scopeMap is the mapping node holding a scope's answers, or nil.
func (a *Answers) scopeMap(s Scope) *yaml.Node {
	section := lookup(a.root(), s.Kind)
	if s.Kind == ScopeGlobal {
		return section
	}
	return lookup(section, s.Name)
}

// pairs yields the key and value nodes of a mapping; anything else
// (nil, a null) yields nothing.
func pairs(m *yaml.Node) func(yield func(k, v *yaml.Node) bool) {
	return func(yield func(k, v *yaml.Node) bool) {
		if m == nil || m.Kind != yaml.MappingNode {
			return
		}
		for i := 0; i+1 < len(m.Content); i += 2 {
			if !yield(m.Content[i], m.Content[i+1]) {
				return
			}
		}
	}
}

// lookup returns the value node for key in mapping m, or nil.
func lookup(m *yaml.Node, key string) *yaml.Node {
	for k, v := range pairs(m) {
		if k.Value == key {
			return v
		}
	}
	return nil
}
