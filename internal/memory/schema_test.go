package memory

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

// TestAnswersSchemaScopes keeps schemas/answers.schema.json's top-level
// keys equal to the three scopes ParseAnswers accepts.
func TestAnswersSchemaScopes(t *testing.T) {
	data, err := os.ReadFile("../../schemas/answers.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var s struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	var got []string
	for k := range s.Properties {
		got = append(got, k)
	}
	want := []string{ScopeCompany, ScopeGlobal, ScopePosting}
	if slices.Sort(got); !slices.Equal(got, want) {
		t.Errorf("schema scopes %v, want %v", got, want)
	}
}
