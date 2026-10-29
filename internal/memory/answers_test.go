package memory

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const answersFixture = "../../testdata/config/answers.fixture.yml"

func loadFixture(t *testing.T) *Answers {
	t.Helper()
	a, err := LoadAnswers(answersFixture)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestParseScope(t *testing.T) {
	good := map[string]Scope{
		"global":            {Kind: ScopeGlobal},
		"company:acme":      {Kind: ScopeCompany, Name: "acme"},
		"posting:a-at-acme": {Kind: ScopePosting, Name: "a-at-acme"},
	}
	for in, want := range good {
		got, err := ParseScope(in)
		if err != nil || got != want || got.String() != in {
			t.Errorf("ParseScope(%q) = %+v, %v", in, got, err)
		}
	}
	for _, in := range []string{"", "global:x", "company", "company: ", "posting:", "role:x", "Global"} {
		if _, err := ParseScope(in); err == nil {
			t.Errorf("ParseScope(%q): no error", in)
		}
	}
}

// start_date is set at all three levels, why_org at two (global and the
// company), lead_story and notice_period at one; referees is empty.
func TestAnswersLookupOrder(t *testing.T) {
	a := loadFixture(t)
	g := func(k, v string) Answer { return Answer{Key: k, Value: v, Scope: "global"} }
	right := g("right_to_work", "EU citizen's spouse permit, no sponsorship needed")
	cases := []struct {
		slug, company string
		want          []Answer
	}{
		{"senior-data-scientist-at-acme", "acme", []Answer{
			{"lead_story", "The fictional trial where I cut\nthe analysis time in half.", "posting:senior-data-scientist-at-acme"},
			g("notice_period", "1 month"), right, g("salary_expectation_eur", "85000"),
			{"start_date", "2027-01-01", "posting:senior-data-scientist-at-acme"},
			{"why_org", "Acme's survival-analysis work matches my research.", "company:acme"},
		}},
		{"other-role-at-acme", "acme", []Answer{
			g("notice_period", "1 month"), right, g("salary_expectation_eur", "85000"),
			{"start_date", "2027-02-01", "company:acme"},
			{"why_org", "Acme's survival-analysis work matches my research.", "company:acme"},
		}},
		{"", "", []Answer{
			g("notice_period", "1 month"), right, g("salary_expectation_eur", "85000"),
			g("start_date", "2027-03-01"), g("why_org", "Generic: I like mission-driven teams."),
		}},
	}
	for _, c := range cases {
		if got := a.For(c.slug, c.company); !reflect.DeepEqual(got, c.want) {
			t.Errorf("For(%q, %q):\n got %+v\nwant %+v", c.slug, c.company, got, c.want)
		}
	}
	form := FormAnswers(a.For("senior-data-scientist-at-acme", "acme"))
	if len(form) != 4 || form[3].Key != "start_date" {
		t.Errorf("FormAnswers kept %+v", form)
	}
}

func TestAnswersAllInFileOrder(t *testing.T) {
	var got []string
	for _, ans := range loadFixture(t).All() {
		got = append(got, ans.Scope+" "+ans.Key)
	}
	want := []string{"global notice_period", "global salary_expectation_eur", "global start_date",
		"global right_to_work", "global why_org", "global referees",
		"company:acme start_date", "company:acme why_org", "company:who why_org",
		"posting:senior-data-scientist-at-acme start_date", "posting:senior-data-scientist-at-acme lead_story"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("All:\n got %v\nwant %v", got, want)
	}
}

// Set then Save keeps every key and every hand comment of the fixture.
func TestAnswersRoundTrip(t *testing.T) {
	a := loadFixture(t)
	before := a.All()
	sets := []struct {
		scope, key, text string
		replaced         bool
	}{
		{"global", "referees", "Dr. A. Example, example.org", true},
		{"posting:senior-data-scientist-at-acme", "start_date", "2027-01-15", true},
		{"global", "why_role", "Line one\nline two", false},
		{"company:globex", "why_org", "Globex is fictional.", false},
	}
	for _, s := range sets {
		scope, _ := ParseScope(s.scope)
		if got := a.Set(scope, s.key, s.text); got != s.replaced {
			t.Errorf("Set(%s, %s) replaced = %v", s.scope, s.key, got)
		}
	}
	path := filepath.Join(t.TempDir(), "answers.yml")
	if err := a.Save(path); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(path); err != nil || fi.Mode().Perm() != 0o600 {
		t.Errorf("mode: %v %v", fi, err)
	}
	saved, _ := os.ReadFile(path)
	keepsComments(t, string(saved))
	b, err := LoadAnswers(path)
	if err != nil {
		t.Fatal(err)
	}
	after := b.All()
	if len(after) != len(before)+2 {
		t.Errorf("%d answers after, want %d", len(after), len(before)+2)
	}
	keepsOrder(t, before, after)
	if got := b.For("", ""); got[len(got)-1] != (Answer{"why_role", "Line one\nline two", "global"}) {
		t.Errorf("multi-line answer: %+v", got)
	}
	if got := b.For("x", "globex"); got[len(got)-2].Scope != "company:globex" {
		t.Errorf("new company scope: %+v", got)
	}
}

// keepsComments checks that every comment of the fixture is in saved.
func keepsComments(t *testing.T, saved string) {
	t.Helper()
	src, _ := os.ReadFile(answersFixture)
	for _, line := range strings.Split(string(src), "\n") {
		if i := strings.Index(line, "#"); i >= 0 && !strings.Contains(saved, line[i:]) {
			t.Errorf("comment lost: %q\n%s", line[i:], saved)
		}
	}
}

// keepsOrder checks that every answer of before is in after, in order.
func keepsOrder(t *testing.T, before, after []Answer) {
	t.Helper()
	j := 0
	for _, ans := range after {
		if j < len(before) && ans.Key == before[j].Key && ans.Scope == before[j].Scope {
			j++
		}
	}
	if j != len(before) {
		t.Errorf("answer %+v lost or moved: %+v", before[j], after)
	}
}

func TestAnswersEmptyFiles(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		"missing.yml": "", "empty.yml": "", "comments.yml": "# kept\n# too\n",
		"null.yml": "global:\ncompany:\n  acme:\n", "tilde.yml": "~\n",
	}
	for name, content := range cases {
		path := filepath.Join(dir, name)
		if name != "missing.yml" {
			_ = os.WriteFile(path, []byte(content), 0o644)
		}
		a, err := LoadAnswers(path)
		if err != nil || len(a.All()) != 0 {
			t.Fatalf("%s: %+v, %v", name, a.All(), err)
		}
		a.Set(Scope{Kind: ScopeCompany, Name: "acme"}, "why_org", "x")
		if err := a.Save(path); err != nil {
			t.Fatal(err)
		}
		b, err := LoadAnswers(path)
		if err != nil || !reflect.DeepEqual(b.All(), []Answer{{"why_org", "x", "company:acme"}}) {
			t.Errorf("%s: %+v, %v", name, b.All(), err)
		}
		if saved, _ := os.ReadFile(path); !strings.HasPrefix(string(saved), "# kept\n# too\n") && name == "comments.yml" {
			t.Errorf("%s: comments lost:\n%s", name, saved)
		}
	}
}

func TestAnswersMalformed(t *testing.T) {
	cases := map[string]string{
		"- a\n":                             "f.yml:1: the top level must be a map",
		"notes: x\n":                        `f.yml:1: unknown top-level key "notes"`,
		"global: [1]\n":                     "f.yml:1: global must be a map of keys",
		"global:\n  a: 1\n  b: [1]\n":       "f.yml:3: global: b must be text",
		"global:\n  a: {x: 1}\n":            "f.yml:2: global: a must be text",
		"global:\n  a: 1\n  a: 2\n":         `f.yml:3: duplicate key "a"`,
		"company: [acme]\n":                 "f.yml:1: company must be a map of names",
		"posting:\n  x: 3\n":                "f.yml:2: posting:x must be a map of keys",
		"global:\n  ? [a]\n  : 1\n":         "f.yml:2: keys must be plain text",
		"global: &g\n  a: 1\nposting: *g\n": "f.yml:3: posting must be a map of names",
		"a: b\n---\nc: d\n":                 "only one YAML document",
		"global: {a: 1\n":                   "f.yml: yaml: line 1",
	}
	for in, want := range cases {
		_, err := ParseAnswers("f.yml", []byte(in))
		if !errors.Is(err, ErrMalformedAnswers) || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: %v, want %q", in, err, want)
		}
	}
}
