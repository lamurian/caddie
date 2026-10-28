package config

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestExamplePreferencesParse(t *testing.T) {
	p, err := LoadPreferences("../../preferences.example.yml")
	if err != nil {
		t.Fatal(err)
	}
	if !p.Apply.RequireGate || p.Ranking.TopN != 10 || p.Salary.Currency != "EUR" {
		t.Errorf("unexpected values: %+v", p)
	}
	if p.Contact.Email == "" || len(p.Sources.Portals) == 0 {
		t.Errorf("contact or portals not parsed: %+v", p)
	}
}

func TestPreferencesDefaultsWhenOmitted(t *testing.T) {
	p, err := ParsePreferences("p.yml", []byte("salary: {currency: EUR}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !p.Apply.RequireGate || p.Apply.FollowUpDays != 21 || p.Apply.DeadlineWarningDays != 7 {
		t.Errorf("apply defaults not applied: %+v", p.Apply)
	}
}

func TestPreferencesExplicitGateFalseIsKept(t *testing.T) {
	p, err := ParsePreferences("p.yml", []byte("apply: {require_gate: false}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Apply.RequireGate {
		t.Error("require_gate: false must be reported, not overridden")
	}
}

func TestPreferencesErrors(t *testing.T) {
	cases := map[string]string{
		"apply:\n  require_gat: true\n": "p.yml: yaml: unmarshal errors:\n  line 2",
		"work_mode: [remote]\n":         "work_mode",
		"salary: {period: week}\n":      "salary.period",
		"ranking: {top_n: 0}\n":         "top_n",
		"salary: [\n":                   "p.yml: yaml: line",
	}
	for input, want := range cases {
		_, err := ParsePreferences("p.yml", []byte(input))
		if err == nil || !strings.Contains(err.Error(), want) || !errors.Is(err, ErrConfig) {
			t.Errorf("%q: got %v, want error containing %q", input, err, want)
		}
	}
}

func TestPreferencesMissingFile(t *testing.T) {
	_, err := LoadPreferences(filepath.Join(t.TempDir(), "preferences.yml"))
	if !errors.Is(err, ErrNoPreferences) || !errors.Is(err, ErrConfig) {
		t.Errorf("got %v", err)
	}
}

func TestContactNeverInJSON(t *testing.T) {
	p := DefaultPreferences()
	p.Contact = Contact{Email: "someone@example.org", Phone: "+00 000"}
	out, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "example.org") || strings.Contains(string(out), "+00") {
		t.Errorf("contact leaked: %s", out)
	}
}
