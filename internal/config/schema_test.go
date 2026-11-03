package config

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
)

// schemaNode is the part of a JSON Schema node this test compares.
type schemaNode struct {
	Ref        string                `json:"$ref"`
	Properties map[string]schemaNode `json:"properties"`
	Items      *schemaNode           `json:"items"`
	Enum       []string              `json:"enum"`
	Defs       map[string]schemaNode `json:"$defs"`
}

func loadSchema(t *testing.T) schemaNode {
	t.Helper()
	data, err := os.ReadFile("../../schemas/preferences.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var s schemaNode
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	return s
}

// resolve follows a local "#/$defs/<name>" reference.
func resolve(root, n schemaNode) schemaNode {
	if name, ok := strings.CutPrefix(n.Ref, "#/$defs/"); ok {
		return root.Defs[name]
	}
	return n
}

// TestSchemaMatchesPreferences keeps schemas/preferences.schema.json and
// the Preferences struct in step: every object has exactly the struct's
// yaml keys, so editors and caddie validate agree on what is allowed.
func TestSchemaMatchesPreferences(t *testing.T) {
	root := loadSchema(t)
	compareStruct(t, root, root, reflect.TypeFor[Preferences](), "preferences")
}

func compareStruct(t *testing.T, root, n schemaNode, typ reflect.Type, path string) {
	t.Helper()
	n = resolve(root, n)
	var keys []string
	for f := range typ.Fields() {
		key := strings.Split(f.Tag.Get("yaml"), ",")[0]
		keys = append(keys, key)
		child, ok := n.Properties[key]
		if !ok {
			continue
		}
		if f.Type.Kind() == reflect.Struct {
			compareStruct(t, root, child, f.Type, path+"."+key)
		}
	}
	got := make([]string, 0, len(n.Properties))
	for k := range n.Properties {
		got = append(got, k)
	}
	sort.Strings(keys)
	sort.Strings(got)
	if !slices.Equal(keys, got) {
		t.Errorf("%s: schema keys %v, struct keys %v", path, got, keys)
	}
}

// TestSchemaEnums keeps the schema's enums equal to the Go validator's.
func TestSchemaEnums(t *testing.T) {
	root := loadSchema(t)
	modes := root.Properties["work_mode"].Items.Enum
	period := root.Properties["salary"].Properties["period"].Enum
	for _, c := range []struct {
		name      string
		got, want []string
	}{{"work_mode", modes, workModes}, {"salary.period", period, periods}} {
		got, want := slices.Sorted(slices.Values(c.got)), slices.Sorted(slices.Values(c.want))
		if !slices.Equal(got, want) {
			t.Errorf("%s: schema %v, Go %v", c.name, got, want)
		}
	}
}

func TestPreferencesNegative(t *testing.T) {
	for _, in := range []string{"salary:\n  minimum: -1\n", "apply:\n  follow_up_days: -2\n"} {
		if _, err := ParsePreferences("p.yml", []byte(in)); err == nil ||
			!strings.Contains(err.Error(), "must not be negative") {
			t.Errorf("%q: got %v", in, err)
		}
	}
}
