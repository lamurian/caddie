package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/document"
)

const letterSlug = "senior-biostatistician-at-example-health-alliance"

// draftWorkspace is a full workspace with the fixture history and the
// fixture draft copied to <dir>/<slug>.md, whose path it returns.
func draftWorkspace(t *testing.T, fixture, dir, slug string) (map[string]string, string) {
	t.Helper()
	vars := workspaceWith(t, "", nil)
	path := filepath.Join(vars["CADDIE_WORKSPACE"], dir, slug+".md")
	copyFile(t, "../../testdata/documents/"+fixture, path)
	return vars, path
}

// letterPosting saves the Acme posting fixture as the Example Health
// Alliance posting that cover-letter-answers.md applies to.
func letterPosting(t *testing.T, root string) {
	t.Helper()
	data, err := os.ReadFile("../../testdata/postings/" + acmeSlug + ".md")
	if err != nil {
		t.Fatal(err)
	}
	s := strings.NewReplacer("slug: "+acmeSlug, "slug: "+letterSlug,
		"title: Senior Data Scientist", "title: Senior Biostatistician",
		"company: Acme", "company: Example Health Alliance").Replace(string(data))
	if err := os.WriteFile(filepath.Join(root, "postings", letterSlug+".md"), []byte(s), 0o600); err != nil {
		t.Fatal(err)
	}
}

// letterAnswers holds the letter's facts under the posting's company and
// posting scopes (the posting's start date beats the company's) when
// scoped is set; otherwise answers.yml has only an unrelated global key.
func letterAnswers(t *testing.T, root string, scoped bool) {
	t.Helper()
	yml := "global:\n  notice_period: 1 month\n"
	if scoped {
		yml += "company:\n  example-health-alliance:\n" +
			"    lead_story: In 2021 I advised the Bergen Outbreak Taskforce on cohort data.\n" +
			"    start_date: \"2027-03-01\"\n" +
			"posting:\n  " + letterSlug + ":\n" +
			"    start_date: \"2027-01-01\"\n"
	}
	if err := os.WriteFile(filepath.Join(root, "answers.yml"), []byte(yml), 0o600); err != nil {
		t.Fatal(err)
	}
}

func verifyJSON(t *testing.T, vars map[string]string, path string) (verifyOutput, result) {
	t.Helper()
	r := runCaddie(t, vars, "verify", "--json", path)
	var out verifyOutput
	if err := json.Unmarshal([]byte(r.stdout), &out); err != nil {
		t.Fatalf("%v: exit %d: %s%s", err, r.code, r.stdout, r.stderr)
	}
	return out, r
}

func TestVerifyPassRecordsHash(t *testing.T) {
	vars, path := draftWorkspace(t, "resume-traceable.md", "resumes", acmeSlug)
	copyFile(t, "../../testdata/postings/"+acmeSlug+".md",
		filepath.Join(vars["CADDIE_WORKSPACE"], "postings", acmeSlug+".md"))
	content, _ := os.ReadFile(path)
	hash := document.Hash(content)
	r := runCaddie(t, vars, "verify", path)
	if r.code != exitOK || r.stderr != "" || strings.Count(r.stdout, "\n") != 1 ||
		!strings.HasPrefix(r.stdout, "verified "+path+": ") || !strings.HasSuffix(r.stdout, " hash "+hash+"\n") {
		t.Fatalf("exit %d: %q %q", r.code, r.stdout, r.stderr)
	}
	state := filepath.Join(vars["CADDIE_WORKSPACE"], ".caddie")
	if _, err := os.Stat(filepath.Join(state, document.VerifiedDir, hash)); err != nil {
		t.Errorf("no hash file: %v", err)
	}
	out, r := verifyJSON(t, vars, path)
	if r.code != exitOK || !out.OK || out.Hash != hash || out.Claims < 30 ||
		out.Untraceable == nil || out.Notices == nil || len(out.Notices) != 0 {
		t.Errorf("--json: %+v", out)
	}
}

// Editing a verified draft, even by one byte, needs a new verify.
func TestVerifyEditInvalidatesHash(t *testing.T) {
	vars, path := draftWorkspace(t, "resume-traceable.md", "resumes", acmeSlug)
	state := filepath.Join(vars["CADDIE_WORKSPACE"], ".caddie")
	if r := runCaddie(t, vars, "verify", path); r.code != exitOK {
		t.Fatalf("exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	content, _ := os.ReadFile(path)
	edited := append(content[:len(content):len(content)], ' ')
	if err := os.WriteFile(path, edited, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		data []byte
		want bool
	}{{content, true}, {edited, false}} {
		if ok, err := document.IsVerified(state, c.data); ok != c.want || err != nil {
			t.Errorf("before re-verify: IsVerified = %v, %v; want %v", ok, err, c.want)
		}
	}
	if r := runCaddie(t, vars, "verify", path); r.code != exitOK {
		t.Fatalf("re-verify: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	if ok, err := document.IsVerified(state, edited); !ok || err != nil {
		t.Errorf("after re-verify: IsVerified = %v, %v", ok, err)
	}
}

func TestVerifyInventedClaims(t *testing.T) {
	cases := []struct {
		fixture string
		want    []document.Claim
	}{
		{"resume-invented-employer.md", []document.Claim{
			{Line: 8, Kind: document.KindEntry, Text: "Acme Pharma Analytics"},
			{Line: 14, Kind: document.KindTerm, Text: "Fjordline Clinical Institute"}}},
		{"resume-invented-skill.md", []document.Claim{
			{Line: 3, Kind: document.KindTerm, Text: "PyTorch"},
			{Line: 4, Kind: document.KindTerm, Text: "Kubernetes"},
			{Line: 8, Kind: document.KindSkill, Text: "deep learning"},
			{Line: 9, Kind: document.KindSkill, Text: "Kubernetes"}}},
	}
	for _, c := range cases {
		vars, path := draftWorkspace(t, c.fixture, "resumes", acmeSlug)
		r := runCaddie(t, vars, "verify", path)
		var want strings.Builder
		for _, cl := range c.want {
			fmt.Fprintf(&want, "%s:%d: %s: %s\n", path, cl.Line, cl.Kind, cl.Text)
		}
		if r.code != exitCheckFailed || !strings.HasPrefix(r.stdout, want.String()) ||
			!strings.Contains(r.stderr, fmt.Sprintf("caddie verify: check failed: %d untraceable", len(c.want))) {
			t.Errorf("%s: exit %d: %s%s", c.fixture, r.code, r.stdout, r.stderr)
		}
		out, r := verifyJSON(t, vars, path)
		if r.code != exitCheckFailed || out.OK || out.Hash != "" || !reflect.DeepEqual(out.Untraceable, c.want) {
			t.Errorf("%s --json: %+v", c.fixture, out)
		}
		if ents, _ := os.ReadDir(filepath.Join(vars["CADDIE_WORKSPACE"], ".caddie", "verified")); len(ents) != 0 {
			t.Errorf("%s: a failed verify recorded a hash", c.fixture)
		}
	}
}

// The letter's facts come from answers scoped to its posting and company.
func TestVerifyCoverLetterAnswers(t *testing.T) {
	for _, scoped := range []bool{true, false} {
		vars, path := draftWorkspace(t, "cover-letter-answers.md", "cover-letters", letterSlug)
		letterPosting(t, vars["CADDIE_WORKSPACE"])
		letterAnswers(t, vars["CADDIE_WORKSPACE"], scoped)
		out, r := verifyJSON(t, vars, path)
		if scoped && (r.code != exitOK || !out.OK) {
			t.Errorf("with answers: exit %d: %+v", r.code, out)
		}
		want := []document.Claim{{Line: 7, Kind: document.KindTerm, Text: "Bergen Outbreak Taskforce"},
			{Line: 12, Kind: document.KindDate, Text: "1 January 2027"}}
		if !scoped && (r.code != exitCheckFailed || !reflect.DeepEqual(out.Untraceable, want)) {
			t.Errorf("without answers: exit %d: %+v", r.code, out)
		}
	}
}

// Without a posting, only global answers count and a notice says so.
func TestVerifyWithoutPosting(t *testing.T) {
	vars, path := draftWorkspace(t, "cover-letter-answers.md", "cover-letters", letterSlug)
	letterAnswers(t, vars["CADDIE_WORKSPACE"], true)
	note := `no posting "` + letterSlug + `": checked against global answers only`
	r := runCaddie(t, vars, "verify", path)
	if r.code != exitCheckFailed || !strings.Contains(r.stdout, "\nnote: "+note+"\n") {
		t.Errorf("exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	out, _ := verifyJSON(t, vars, path)
	if !reflect.DeepEqual(out.Notices, []string{note}) || len(out.Untraceable) < 2 {
		t.Errorf("--json: %+v", out)
	}
	vars, path = draftWorkspace(t, "resume-traceable.md", "resumes", "My Draft")
	if out, r = verifyJSON(t, vars, path); r.code != exitOK || len(out.Notices) != 1 {
		t.Errorf("not a slug: exit %d: %+v", r.code, out)
	}
}

func TestVerifyErrors(t *testing.T) {
	vars, path := draftWorkspace(t, "resume-traceable.md", "resumes", acmeSlug)
	missing := filepath.Join(filepath.Dir(path), "absent.md")
	cases := []struct {
		args []string
		code int
		want string
	}{
		{[]string{"verify"}, exitUsage, "verify <doc.md>"},
		{[]string{"verify", path, path}, exitUsage, "verify <doc.md>"},
		{[]string{"verify", "--bogus", path}, exitUsage, "flag provided but not defined"},
		{[]string{"verify", missing}, exitNotFound, `no file "` + missing + `"`},
	}
	for _, c := range cases {
		r := runCaddie(t, vars, c.args...)
		if r.code != c.code || !strings.Contains(r.stderr, c.want) {
			t.Errorf("%v: exit %d: %s", c.args, r.code, r.stderr)
		}
	}
	bad := workspaceWith(t, "", map[string]string{"teaching.md": "## Teaching\n### TA\nUniversity\nsometime\n"})
	if r := runCaddie(t, bad, "verify", path); r.code != exitMalformed || !strings.Contains(r.stderr, "teaching.md:4") {
		t.Errorf("malformed history: exit %d: %s", r.code, r.stderr)
	}
}
