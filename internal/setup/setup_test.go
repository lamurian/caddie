package setup

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/memory"
)

var portals = []string{"who", "un", "bcg"}

func TestParse(t *testing.T) {
	qs := byID(Questions(portals))
	cases := []struct {
		id, raw string
		want    any
		bad     bool
	}{
		{"salary.minimum", "70000", 70000, false},
		{"salary.minimum", "", 0, false},
		{"salary.minimum", "-5", nil, true},
		{"salary.period", "month", "month", false},
		{"salary.period", "week", nil, true},
		{"roles.include", " data scientist,, epidemiologist ", []string{"data scientist", "epidemiologist"}, false},
		{"roles.include", "", nil, true},
		{"work_mode", "wfa, wfh", []string{"wfa", "wfh"}, false},
		{"work_mode", "wfh, wfh", nil, true},
		{"work_mode", "office", nil, true},
		{"sources.portals", "", []string{"who", "un", "bcg"}, false},
		{"global.referees", "", "Available on request", false},
	}
	for _, c := range cases {
		got, err := Parse(qs[c.id], c.raw)
		if c.bad != (err != nil) || (!c.bad && !reflect.DeepEqual(got, c.want)) {
			t.Errorf("%s %q: got %#v, %v", c.id, c.raw, got, err)
		}
		if err != nil && !errors.Is(err, ErrAnswer) {
			t.Errorf("%s: error %v does not wrap ErrAnswer", c.id, err)
		}
	}
}

func byID(qs []Question) map[string]Question {
	m := map[string]Question{}
	for _, q := range qs {
		m[q.ID] = q
	}
	return m
}

func TestAskBuildsValidFiles(t *testing.T) {
	qs := Questions(portals)
	in := strings.Join([]string{"me@example.org", "", "data scientist, biostatistician", "",
		"", "weekly", "year", "70000", "85000", "wfa, hybrid", "Groningen, Remote (EU)"}, "\n") + "\n"
	var out strings.Builder
	v, err := Ask(qs, strings.NewReader(in), &out)
	if err != nil {
		t.Fatalf("ask: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), `"weekly" is not one of year|month|hour`) {
		t.Errorf("no re-ask after a bad choice:\n%s", out.String())
	}
	checkPrefs(t, v)
	ans, err := Answers(v)
	if err != nil {
		t.Fatal(err)
	}
	a, err := memory.ParseAnswers("a.yml", ans)
	if err != nil || len(a.All()) != 1 || a.All()[0].Key != "referees" {
		t.Errorf("answers: %v\n%s", err, ans)
	}
}

func checkPrefs(t *testing.T, v Values) {
	t.Helper()
	prefs, err := Preferences(v)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(prefs), "# yaml-language-server: $schema=schemas/preferences.schema.json\n") {
		t.Errorf("no schema header:\n%s", prefs)
	}
	p, err := config.ParsePreferences("p.yml", prefs)
	if err != nil {
		t.Fatalf("%v\n%s", err, prefs)
	}
	if !p.Apply.RequireGate || p.Salary.Minimum != 70000 || p.Contact.Email != "me@example.org" ||
		!reflect.DeepEqual(p.WorkMode, []string{"wfa", "hybrid"}) || p.Sources.Adzuna.Enabled ||
		!reflect.DeepEqual(p.Sources.Portals, portals) || p.Ranking.TopN != 10 || p.Apply.FollowUpDays != 21 {
		t.Errorf("preferences: %+v", p)
	}
}

func TestAskRequiredAtEOF(t *testing.T) {
	if _, err := Ask(Questions(portals), strings.NewReader("me@example.org\n\n"), &strings.Builder{}); !errors.Is(err, ErrAnswer) {
		t.Errorf("roles.include unanswered at EOF: got %v", err)
	}
}

func TestFromJSON(t *testing.T) {
	qs := Questions(portals)
	v, err := FromJSON(qs, map[string]any{
		"roles.include": []any{"epidemiologist"}, "salary.minimum": float64(50000),
		"sources.adzuna.countries": "nl, de", "global.notice_period": "1 month",
	})
	if err != nil {
		t.Fatal(err)
	}
	prefs, err := Preferences(v)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := config.ParsePreferences("p.yml", prefs)
	if !p.Sources.Adzuna.Enabled || p.Salary.Minimum != 50000 || p.Roles.Include[0] != "epidemiologist" {
		t.Errorf("preferences: %+v", p)
	}
	for _, bad := range []map[string]any{
		{"roles.include": "x", "role.include": "typo"},
		{"roles.include": "x", "salary.minimum": 1.5},
		{"roles.include": []any{1}},
		{"roles.include": "x", "work_mode": "office"},
	} {
		if _, err := FromJSON(qs, bad); !errors.Is(err, ErrAnswer) {
			t.Errorf("%v: got %v", bad, err)
		}
	}
}

func TestChecksAndWrite(t *testing.T) {
	dir := t.TempDir()
	prefs, answers := filepath.Join(dir, "preferences.yml"), filepath.Join(dir, "answers.yml")
	if err := CheckPreferences(prefs); !errors.Is(err, ErrMissing) || !errors.Is(err, config.ErrConfig) {
		t.Errorf("missing preferences: %v", err)
	}
	if err := CheckAnswers(answers); !errors.Is(err, ErrMissing) {
		t.Errorf("missing answers: %v", err)
	}
	if err := Write(prefs, []byte("apply:\n  require_gate: false\n")); err != nil {
		t.Fatal(err)
	}
	if err := Write(prefs, []byte("x")); err == nil {
		t.Error("Write overwrote an existing file")
	}
	if err := CheckPreferences(prefs); !errors.Is(err, config.ErrConfig) || !strings.Contains(err.Error(), "require_gate") {
		t.Errorf("gate off: %v", err)
	}
	if err := os.WriteFile(answers, []byte("global:\n  k: [a]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := CheckAnswers(answers); !errors.Is(err, memory.ErrMalformedAnswers) {
		t.Errorf("malformed answers: %v", err)
	}
	if ok, err := CopyIfMissing(answers, filepath.Join(dir, "copy.yml")); !ok || err != nil {
		t.Errorf("copy: %v %v", ok, err)
	}
	if ok, _ := CopyIfMissing(answers, prefs); ok {
		t.Error("CopyIfMissing replaced an existing file")
	}
}
