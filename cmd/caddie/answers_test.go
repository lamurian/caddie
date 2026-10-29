package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/memory"
)

const acmeSlug = "senior-data-scientist-at-acme"

// answersWorkspace is a workspace holding the answers fixture as
// answers.yml and the Acme posting in dir (postings, manual or applied).
func answersWorkspace(t *testing.T, dir string) map[string]string {
	t.Helper()
	vars := newWorkspace(t, "")
	root := vars["CADDIE_WORKSPACE"]
	copyFile(t, "../../testdata/config/answers.fixture.yml", filepath.Join(root, "answers.yml"))
	if dir != "" {
		copyFile(t, "../../testdata/postings/"+acmeSlug+".md", filepath.Join(root, dir, acmeSlug+".md"))
	}
	return vars
}

func copyFile(t *testing.T, from, to string) {
	t.Helper()
	data, err := os.ReadFile(from)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(to, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func showAnswers(t *testing.T, vars map[string]string, args ...string) showOutput {
	t.Helper()
	r := runCaddie(t, vars, append([]string{"answers"}, args...)...)
	var out showOutput
	if r.code != exitOK || json.Unmarshal([]byte(r.stdout), &out) != nil {
		t.Fatalf("%v: exit %d: %s%s", args, r.code, r.stdout, r.stderr)
	}
	return out
}

func keysAndScopes(list []memory.Answer) []string {
	out := []string{}
	for _, a := range list {
		out = append(out, a.Key+"@"+a.Scope)
	}
	return out
}

func TestAnswersAddCommand(t *testing.T) {
	vars := newWorkspace(t, "")
	path := filepath.Join(vars["CADDIE_WORKSPACE"], "answers.yml")
	r := runCaddie(t, vars, "answers", "add", "--key", "notice_period", "--scope", "global", "1 month")
	want := "{\n  \"scope\": \"global\",\n  \"key\": \"notice_period\",\n  \"value\": \"1 month\",\n  \"replaced\": false\n}\n"
	if r.code != exitOK || r.stdout != want {
		t.Fatalf("new: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	r = runCaddie(t, vars, "answers", "add", " 2 months ", "--scope", "global", "--key", "notice_period")
	if r.code != exitOK || !strings.Contains(r.stdout, `"value": "2 months",`) ||
		!strings.Contains(r.stdout, `"replaced": true`) {
		t.Errorf("replace: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	r = runCaddie(t, vars, "answers", "add", "--key", "why_org", "--scope", "company:Société Générale", "Banks.")
	if r.code != exitOK || !strings.Contains(r.stdout, `"scope": "company:societe-generale"`) {
		t.Errorf("company: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	if fi, err := os.Stat(path); err != nil || fi.Mode().Perm() != 0o600 {
		t.Errorf("answers.yml mode: %v %v", fi, err)
	}
	got := keysAndScopes(showAnswers(t, vars, "list").Answers)
	if want := []string{"notice_period@global", "why_org@company:societe-generale"}; !reflect.DeepEqual(got, want) {
		t.Errorf("list = %v, want %v", got, want)
	}
}

func TestAnswersShowCommand(t *testing.T) {
	vars := answersWorkspace(t, "postings")
	out := showAnswers(t, vars, "show")
	want := []string{"notice_period@global", "right_to_work@global", "salary_expectation_eur@global",
		"start_date@global", "why_org@global"}
	if got := keysAndScopes(out.Answers); out.For != "" || !reflect.DeepEqual(got, want) {
		t.Errorf("show = %+v", out)
	}
	merged := []string{"lead_story@posting:" + acmeSlug, "notice_period@global", "right_to_work@global",
		"salary_expectation_eur@global", "start_date@posting:" + acmeSlug, "why_org@company:acme"}
	out = showAnswers(t, vars, "show", "--for", acmeSlug, "--json")
	if got := keysAndScopes(out.Answers); out.For != acmeSlug || out.Company != "acme" || !reflect.DeepEqual(got, merged) {
		t.Errorf("show --for = %+v", out)
	}
	form := []string{"notice_period@global", "right_to_work@global", "salary_expectation_eur@global",
		"start_date@posting:" + acmeSlug}
	if got := keysAndScopes(showAnswers(t, vars, "show", "--form", "--for", acmeSlug).Answers); !reflect.DeepEqual(got, form) {
		t.Errorf("show --form = %v", got)
	}
	// manual/ and applied/ are searched when postings/ has no such slug.
	for _, dir := range []string{"manual", "applied"} {
		v := answersWorkspace(t, dir)
		if got := keysAndScopes(showAnswers(t, v, "show", "--for", acmeSlug).Answers); !reflect.DeepEqual(got, merged) {
			t.Errorf("%s: show --for = %v", dir, got)
		}
	}
	r := runCaddie(t, vars, "answers", "show", "--for", "no-such-posting")
	if r.code != exitNotFound || !strings.Contains(r.stderr, `no posting "no-such-posting"`) {
		t.Errorf("unknown slug: %+v", r)
	}
}

func TestAnswersListCommand(t *testing.T) {
	got := showAnswers(t, answersWorkspace(t, ""), "list").Answers
	if len(got) != 11 || got[5] != (memory.Answer{Key: "referees", Value: "", Scope: "global"}) ||
		got[10].Scope != "posting:"+acmeSlug {
		t.Errorf("list = %+v", got)
	}
}

func TestAnswersMalformedFile(t *testing.T) {
	vars := answersWorkspace(t, "postings")
	path := filepath.Join(vars["CADDIE_WORKSPACE"], "answers.yml")
	if err := os.WriteFile(path, []byte("global:\n  salary: [1, 2]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"list"}, {"show"}, {"show", "--for", acmeSlug},
		{"add", "--key", "k", "--scope", "global", "v"},
	} {
		r := runCaddie(t, vars, append([]string{"answers"}, args...)...)
		if r.code != exitMalformed || !strings.Contains(r.stderr, path+":2: global: salary must be text") {
			t.Errorf("%v: %+v", args, r)
		}
	}
}

func TestAnswersUsage(t *testing.T) {
	vars := newWorkspace(t, "")
	add := []string{"answers", "add", "--key", "k", "--scope"}
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"answers"}, "answers add --key k"},
		{[]string{"answers", "forget"}, `unknown subcommand "forget"`},
		{[]string{"answers", "add", "--key", "Bad-Key", "--scope", "global", "x"}, `--key "Bad-Key"`},
		{[]string{"answers", "add", "--scope", "global", "x"}, `--key ""`},
		{append(add, "role:x", "x"), `--scope: scope "role:x"`},
		{append(add, "company:", "x"), `--scope: scope "company:"`},
		{append(add, "company:!!", "x"), "no letters or digits"},
		{append(add, "posting:Not A Slug", "x"), `"Not A Slug" is not a posting slug`},
		{append(add, "global"), "exactly one answer text"},
		{append(add, "global", "a", "b"), "exactly one answer text"},
		{append(add, "global", "  "), "answer text is empty"},
		{[]string{"answers", "show", "--for", "../x"}, `"../x" is not a posting slug`},
		{[]string{"answers", "show", "extra"}, `unexpected argument "extra"`},
		{[]string{"answers", "list", "--bogus"}, "flag provided but not defined"},
	}
	for _, c := range cases {
		r := runCaddie(t, vars, c.args...)
		if r.code != exitUsage || !strings.Contains(r.stderr, c.want) {
			t.Errorf("%v: code %d, %q; want 2 and %q", c.args, r.code, r.stderr, c.want)
		}
	}
	if _, err := os.Stat(filepath.Join(vars["CADDIE_WORKSPACE"], "answers.yml")); !os.IsNotExist(err) {
		t.Errorf("a usage error wrote answers.yml: %v", err)
	}
}
