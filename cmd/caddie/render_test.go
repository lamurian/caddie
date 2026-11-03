package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/document"
)

const (
	resumePDF = "Mira-Example-Halvorsen-resume-senior-data-scientist-acme.pdf"
	letterPDF = "Mira-Example-Halvorsen-cover-letter-senior-data-scientist-acme.pdf"
)

// fakeQuarto records each run and writes the PDF named by --output, or
// fails when fail is set.
type fakeQuarto struct {
	calls [][]string
	json  string // _caddie.json of the last run
	fail  bool
}

func (f *fakeQuarto) Run(_ context.Context, dir string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	if data, err := os.ReadFile(filepath.Join(dir, "_caddie.json")); err == nil {
		f.json = string(data)
	}
	if f.fail {
		return []byte("ERROR: compile failed"), errors.New("exit status 1")
	}
	i := slices.Index(args, "--output")
	return nil, os.WriteFile(filepath.Join(dir, args[i+1]), []byte("%PDF-fake"), 0o600)
}

// noQuarto is the quarto of every test that must not render.
type noQuarto struct{ t *testing.T }

func (n noQuarto) Run(context.Context, string, ...string) ([]byte, error) {
	n.t.Error("quarto ran")
	return nil, errors.New("no quarto in tests")
}

// renderWorkspace holds the fixture history and preferences, both
// templates, both Acme drafts and the Acme posting with the given status
// in postings/.
func renderWorkspace(t *testing.T, status string) (map[string]string, string) {
	t.Helper()
	prefs, err := os.ReadFile("../../testdata/config/preferences.fixture.yml")
	if err != nil {
		t.Fatal(err)
	}
	vars := workspaceWith(t, string(prefs), nil)
	root := vars["CADDIE_WORKSPACE"]
	copyFile(t, "../../testdata/documents/resume-traceable.md", filepath.Join(root, "resumes", acmeSlug+".md"))
	copyFile(t, "../../testdata/documents/cover-letter-answers.md", filepath.Join(root, "cover-letters", acmeSlug+".md"))
	for _, f := range []string{"resume.qmd", "cover-letter.qmd"} {
		writeFile(t, filepath.Join(root, "templates", f), "---\nformat: typst\n---\n")
	}
	data, err := os.ReadFile("../../testdata/postings/" + acmeSlug + ".md")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, postingFile(root), strings.Replace(string(data), "status: saved", "status: "+status, 1))
	return vars, root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func postingFile(root string) string { return filepath.Join(root, "postings", acmeSlug+".md") }

// markVerified records the draft's current content as verified.
func markVerified(t *testing.T, root, rel string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := document.RecordVerified(filepath.Join(root, ".caddie"), rel, content, "2026-10-28"); err != nil {
		t.Fatal(err)
	}
}

func postingStatus(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(postingFile(root))
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range strings.Split(string(data), "\n") {
		if s, ok := strings.CutPrefix(l, "status: "); ok {
			return s
		}
	}
	return ""
}

func TestRenderResume(t *testing.T) {
	vars, root := renderWorkspace(t, "saved")
	markVerified(t, root, "resumes/"+acmeSlug+".md")
	body := func() string {
		data, _ := os.ReadFile(postingFile(root))
		_, b, _ := strings.Cut(string(data), "\n---\n")
		return b
	}
	before := body()
	q := &fakeQuarto{}
	r := runQuarto(t, vars, nil, q, "render", "resume", acmeSlug)
	want := "rendered resumes/" + resumePDF + "\nhandoff handoff/" + resumePDF + "\nstatus rendered\n"
	if r.code != exitOK || r.stdout != want || r.stderr != "" {
		t.Fatalf("exit %d: %q %q", r.code, r.stdout, r.stderr)
	}
	wantArgs := []string{"render", "resume.qmd", "--to", "typst", "--output", resumePDF}
	if len(q.calls) != 1 || !slices.Equal(q.calls[0], wantArgs) {
		t.Errorf("quarto args %q, want %q", q.calls, wantArgs)
	}
	if !strings.Contains(q.json, "mira@example.org") || !strings.Contains(q.json, `"Senior Data Scientist"`) {
		t.Errorf("_caddie.json %s", q.json)
	}
	for _, p := range []string{"resumes/" + resumePDF, "handoff/" + resumePDF} {
		if data, err := os.ReadFile(filepath.Join(root, p)); err != nil || string(data) != "%PDF-fake" {
			t.Errorf("%s: %q, %v", p, data, err)
		}
	}
	if s := postingStatus(t, root); s != "rendered" {
		t.Errorf("status %q", s)
	}
	if after := body(); after != before {
		t.Errorf("posting body changed:\n%s", after)
	}
}

// The letter can be rendered after the resume; the status stays rendered.
func TestRenderCoverLetterAfterResume(t *testing.T) {
	vars, root := renderWorkspace(t, "tailored")
	markVerified(t, root, "resumes/"+acmeSlug+".md")
	markVerified(t, root, "cover-letters/"+acmeSlug+".md")
	q := &fakeQuarto{}
	for _, kind := range []string{"resume", "cover-letter"} {
		if r := runQuarto(t, vars, nil, q, "render", kind, acmeSlug); r.code != exitOK {
			t.Fatalf("%s: exit %d: %s%s", kind, r.code, r.stdout, r.stderr)
		}
	}
	r := runQuarto(t, vars, nil, q, "render", "--json", "cover-letter", acmeSlug)
	var out map[string]string
	if err := json.Unmarshal([]byte(r.stdout), &out); err != nil || r.code != exitOK {
		t.Fatalf("exit %d: %v: %s%s", r.code, err, r.stdout, r.stderr)
	}
	want := map[string]string{"kind": "cover-letter", "slug": acmeSlug, "status": "rendered",
		"pdf": "cover-letters/" + letterPDF, "handoff": "handoff/" + letterPDF}
	if len(out) != len(want) {
		t.Errorf("--json %v, want %v", out, want)
	}
	for k, v := range want {
		if out[k] != v {
			t.Errorf("--json %s = %q, want %q", k, out[k], v)
		}
	}
	if s := postingStatus(t, root); s != "rendered" || len(q.calls) != 3 {
		t.Errorf("status %q after %d renders", s, len(q.calls))
	}
}

// The gate: a draft never verified, or edited since, is refused (exit 7)
// before quarto runs, and the status stays put.
func TestRenderRefusesUnverified(t *testing.T) {
	vars, root := renderWorkspace(t, "saved")
	draft := filepath.Join(root, "resumes", acmeSlug+".md")
	r := runCaddie(t, vars, "render", "resume", acmeSlug)
	if r.code != exitRefused || !strings.Contains(r.stderr, "caddie render: not verified") ||
		!strings.Contains(r.stderr, "run caddie verify") {
		t.Errorf("unverified: exit %d: %q", r.code, r.stderr)
	}
	markVerified(t, root, "resumes/"+acmeSlug+".md")
	f, err := os.OpenFile(draft, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("\nOne more line.\n")
	_ = f.Close()
	if r := runCaddie(t, vars, "render", "resume", acmeSlug); r.code != exitRefused {
		t.Errorf("edited: exit %d: %q", r.code, r.stderr)
	}
	if s := postingStatus(t, root); s != "saved" {
		t.Errorf("status %q", s)
	}
}

func TestRenderRefusesPastStatus(t *testing.T) {
	for _, status := range []string{"applied", "manual"} {
		vars, root := renderWorkspace(t, status)
		markVerified(t, root, "resumes/"+acmeSlug+".md")
		if r := runCaddie(t, vars, "render", "resume", acmeSlug); r.code != exitIllegalState {
			t.Errorf("%s: exit %d: %q", status, r.code, r.stderr)
		}
	}
	vars, root := renderWorkspace(t, "applied")
	markVerified(t, root, "resumes/"+acmeSlug+".md")
	if err := os.Rename(postingFile(root), filepath.Join(root, "applied", acmeSlug+".md")); err != nil {
		t.Fatal(err)
	}
	r := runCaddie(t, vars, "render", "resume", acmeSlug)
	if r.code != exitIllegalState || !strings.Contains(r.stderr, "not in postings/") {
		t.Errorf("in applied/: exit %d: %q", r.code, r.stderr)
	}
}

func TestRenderErrors(t *testing.T) {
	cases := []struct {
		name  string
		setup func(root string)
		args  []string
		code  int
		want  string
	}{
		{"no draft", func(root string) { _ = os.Remove(filepath.Join(root, "resumes", acmeSlug+".md")) },
			nil, exitNotFound, "no draft resumes/" + acmeSlug + ".md: write it, then run caddie verify"},
		{"no posting", func(root string) { _ = os.Remove(postingFile(root)) },
			nil, exitNotFound, "no posting postings/" + acmeSlug + ".md"},
		{"no template", func(root string) { _ = os.Remove(filepath.Join(root, "templates", "resume.qmd")) },
			nil, exitConfig, "CADDIE_RESUME_TEMPLATE"},
		{"bad kind", nil, []string{"render", "letter", acmeSlug}, exitUsage, "want resume|cover-letter"},
		{"one arg", nil, []string{"render", "resume"}, exitUsage, "render resume|cover-letter <slug>"},
		{"bad slug", nil, []string{"render", "resume", "../x"}, exitUsage, "not a posting slug"},
	}
	for _, c := range cases {
		vars, root := renderWorkspace(t, "saved")
		markVerified(t, root, "resumes/"+acmeSlug+".md")
		if c.setup != nil {
			c.setup(root)
		}
		args := c.args
		if args == nil {
			args = []string{"render", "resume", acmeSlug}
		}
		r := runCaddie(t, vars, args...)
		if r.code != c.code || !strings.Contains(r.stderr, c.want) {
			t.Errorf("%s: exit %d: %q; want %d and %q", c.name, r.code, r.stderr, c.code, c.want)
		}
	}
}

func TestRenderQuartoFailure(t *testing.T) {
	vars, root := renderWorkspace(t, "saved")
	markVerified(t, root, "resumes/"+acmeSlug+".md")
	r := runQuarto(t, vars, nil, &fakeQuarto{fail: true}, "render", "resume", acmeSlug)
	if r.code != exitExternal || !strings.Contains(r.stderr, "compile failed") {
		t.Errorf("exit %d: %q", r.code, r.stderr)
	}
	if s := postingStatus(t, root); s != "saved" {
		t.Errorf("status %q", s)
	}
	if _, err := os.Stat(filepath.Join(root, "handoff", resumePDF)); err == nil {
		t.Error("handoff copy after a failed render")
	}
}

// Contact reaches only the render directory: `caddie profile` (without
// --with-contact) after a render still holds no e-mail or phone.
func TestRenderKeepsContactOutOfProfile(t *testing.T) {
	vars, root := renderWorkspace(t, "saved")
	markVerified(t, root, "resumes/"+acmeSlug+".md")
	if r := runQuarto(t, vars, nil, &fakeQuarto{}, "render", "resume", acmeSlug); r.code != exitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	r := runCaddie(t, vars, "profile")
	if r.code != exitOK || strings.Contains(r.stdout, "mira@example.org") || strings.Contains(r.stdout, "00000000") {
		t.Errorf("exit %d: profile leaks contact:\n%s", r.code, r.stdout)
	}
}
