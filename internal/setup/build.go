package setup

import (
	"bytes"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/memory"
)

const (
	prefsHeader = "# yaml-language-server: $schema=schemas/preferences.schema.json\n" +
		"# Written by caddie init; edit freely, then run caddie validate.\n"
	answersHeader = "# yaml-language-server: $schema=schemas/answers.schema.json\n" +
		"# Written by caddie init. Caddie adds answers here via `caddie answers add`.\n" +
		"# Lookup order for a key: posting:<slug> -> company:<x> -> global.\n"
)

// Preferences renders preferences.yml from v. Keys no question covers
// take the example file's values; apply.require_gate is always true. The
// result is validated with the same loader every command uses.
func Preferences(v Values) ([]byte, error) {
	doc := map[string]any{
		"sources": map[string]any{
			"remotive":  map[string]any{"enabled": true},
			"arbeitnow": map[string]any{"enabled": true},
		},
		"ranking": map[string]any{"weights": map[string]any{
			"skills_match": 0.4, "salary_fit": 0.2, "work_mode_fit": 0.2, "location_fit": 0.2,
		}},
		"apply": map[string]any{"require_gate": true},
	}
	for _, q := range For(Questions(nil), FilePreferences) {
		if val, ok := v[q.ID]; ok {
			setPath(doc, q.ID, val)
		}
	}
	countries, _ := v["sources.adzuna.countries"].([]string)
	setPath(doc, "sources.adzuna.enabled", len(countries) > 0)
	raw, err := yaml.Marshal(doc)
	if err != nil {
		return nil, err
	}
	p, err := config.ParsePreferences("preferences.yml", raw)
	if err != nil {
		return nil, err
	}
	p.Apply.RequireGate = true
	return encode(prefsHeader, p)
}

// Answers renders answers.yml from the global answers in v; empty ones
// are left out, since an empty value counts as unanswered.
func Answers(v Values) ([]byte, error) {
	global := map[string]string{}
	for id, val := range v {
		key, ok := strings.CutPrefix(id, memory.ScopeGlobal+".")
		if s, isText := val.(string); ok && isText && s != "" {
			global[key] = s
		}
	}
	out, err := encode(answersHeader, map[string]any{memory.ScopeGlobal: global})
	if err != nil {
		return nil, err
	}
	if _, err := memory.ParseAnswers("answers.yml", out); err != nil {
		return nil, err
	}
	return out, nil
}

func encode(header string, v any) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(header)
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// setPath sets a dotted path in nested maps, creating maps on the way.
func setPath(doc map[string]any, path string, val any) {
	keys := strings.Split(path, ".")
	m := doc
	for _, k := range keys[:len(keys)-1] {
		next, ok := m[k].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[k] = next
		}
		m = next
	}
	m[keys[len(keys)-1]] = val
}
