package document

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/profile"
)

const (
	testDraft = "---\ntitle: Acme draft\n---\n# Experience\n\nBuilt models.\n"
	testBody  = "# Experience\n\nBuilt models.\n"
	testEmail = "mira@example.org"
	testPhone = "+00 0 00000000"
	testPDF   = "Mira-Example-Halvorsen-resume-senior-data-scientist-acme-co.pdf"
)

// fakeQuarto records its call and a snapshot of the render directory, then
// writes a fake PDF holding _caddie.json (contact included, as a real one).
type fakeQuarto struct {
	calls int
	dir   string
	args  []string
	files map[string]string
	out   []byte
	err   error
	noPDF bool
}

func (f *fakeQuarto) Run(_ context.Context, dir string, args ...string) ([]byte, error) {
	f.calls++
	f.dir, f.args = dir, args
	f.files = snapshot(dir)
	if f.err != nil || f.noPDF {
		return f.out, f.err
	}
	pdf := "%PDF-fake\n" + f.files[DataFile]
	return f.out, os.WriteFile(filepath.Join(dir, args[len(args)-1]), []byte(pdf), 0o600)
}

// snapshot maps every file under dir (slash paths) to its content.
func snapshot(dir string) map[string]string {
	files := map[string]string{}
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(dir, path)
			data, _ := os.ReadFile(path)
			files[filepath.ToSlash(rel)] = string(data)
		}
		return nil
	})
	return files
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// newRenderFixture lays out a workspace with a template directory and a
// resume draft, verified when verify is true. TMPDIR points at tmp, an
// empty directory outside the workspace.
func newRenderFixture(t *testing.T, verify bool) (in RenderInput, ws, tmp string) {
	t.Helper()
	ws, tmp = t.TempDir(), t.TempDir()
	t.Setenv("TMPDIR", tmp)
	tpl := filepath.Join(ws, "templates")
	writeTestFile(t, filepath.Join(tpl, "resume.qmd"), "{{< include _caddie-body.md >}}\n")
	writeTestFile(t, filepath.Join(tpl, "typst-template.typ"), "#let caddie = json(\"_caddie.json\")\n")
	writeTestFile(t, filepath.Join(tpl, "partials", "style.typ"), "// style\n")
	writeTestFile(t, filepath.Join(tpl, ".hidden"), "skip me\n")
	writeTestFile(t, filepath.Join(tpl, ".git", "config"), "skip me\n")
	src := filepath.Join(ws, "resumes", "acme.md")
	writeTestFile(t, src, testDraft)
	if verify {
		if _, err := RecordVerified(filepath.Join(ws, ".caddie"), "resumes/acme.md", []byte(testDraft), "2026-10-29"); err != nil {
			t.Fatal(err)
		}
	}
	return RenderInput{
		Kind: KindResume, Source: src,
		StateDir:   filepath.Join(ws, ".caddie"),
		HandoffDir: filepath.Join(ws, "handoff"),
		Template:   filepath.Join(tpl, "resume.qmd"),
		Identity: profile.Identity{Name: "Mira Example-Halvorsen", Location: "Utrecht, NL",
			Links: map[string]string{"orcid": "https://orcid.org/0000", "github": "https://github.com/example", "blog": ""}},
		Contact:  config.Contact{Email: testEmail, Phone: testPhone},
		Position: "Senior Data Scientist", Company: "Acme & Co.",
	}, ws, tmp
}

// assertNoOutput checks that no PDF, handoff file or temp dir is left.
func assertNoOutput(t *testing.T, in RenderInput, tmp string) {
	t.Helper()
	if pdfs, _ := filepath.Glob(filepath.Join(filepath.Dir(in.Source), "*.pdf")); len(pdfs) > 0 {
		t.Errorf("PDF left next to the source: %v", pdfs)
	}
	if entries, _ := os.ReadDir(in.HandoffDir); len(entries) > 0 {
		t.Errorf("handoff not empty: %v", entries)
	}
	if entries, _ := os.ReadDir(tmp); len(entries) > 0 {
		t.Errorf("temp render dir left behind: %v", entries)
	}
}

func TestRenderRefusesUnverified(t *testing.T) {
	cases := map[string]func(t *testing.T, in RenderInput){
		"never verified": func(*testing.T, RenderInput) {},
		"edited by one byte": func(t *testing.T, in RenderInput) {
			writeTestFile(t, in.Source, testDraft+" ")
		},
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			in, _, tmp := newRenderFixture(t, name != "never verified")
			edit(t, in)
			q := &fakeQuarto{}
			_, err := Render(context.Background(), q, in)
			if !errors.Is(err, ErrNotVerified) {
				t.Fatalf("err = %v, want ErrNotVerified", err)
			}
			if msg := err.Error(); !strings.Contains(msg, "run caddie verify "+in.Source) {
				t.Errorf("message %q does not name the draft and the fix", msg)
			}
			if q.calls != 0 {
				t.Errorf("quarto called %d times", q.calls)
			}
			assertNoOutput(t, in, tmp)
		})
	}
}

func TestRenderWritesPDFAndHandoff(t *testing.T) {
	for _, existing := range []bool{false, true} {
		in, ws, tmp := newRenderFixture(t, true)
		if existing {
			writeTestFile(t, filepath.Join(in.HandoffDir, testPDF), "old render")
		}
		q := &fakeQuarto{}
		res, err := Render(context.Background(), q, in)
		if err != nil {
			t.Fatal(err)
		}
		checkQuartoCall(t, q, tmp)
		want := RenderResult{PDF: filepath.Join(ws, "resumes", testPDF), Handoff: filepath.Join(ws, "handoff", testPDF)}
		if res != want {
			t.Errorf("result = %+v, want %+v", res, want)
		}
		for _, p := range []string{res.PDF, res.Handoff} {
			if data, err := os.ReadFile(p); err != nil || !strings.HasPrefix(string(data), "%PDF-fake") {
				t.Errorf("%s: %q, %v", p, data, err)
			}
		}
		if _, err := os.Stat(q.dir); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("temp dir %s not removed: %v", q.dir, err)
		}
	}
}

// checkQuartoCall asserts argv, the render dir and what it held.
func checkQuartoCall(t *testing.T, q *fakeQuarto, tmp string) {
	t.Helper()
	wantArgs := []string{"render", "resume.qmd", "--to", "typst", "--output", testPDF}
	if q.calls != 1 || !reflect.DeepEqual(q.args, wantArgs) {
		t.Errorf("quarto calls=%d args=%q, want 1 call %q", q.calls, q.args, wantArgs)
	}
	if filepath.Dir(q.dir) != tmp || !strings.HasPrefix(filepath.Base(q.dir), "caddie-render-") {
		t.Errorf("render dir %s, want caddie-render-* in %s", q.dir, tmp)
	}
	if got := q.files[BodyFile]; got != testBody {
		t.Errorf("%s = %q, want %q", BodyFile, got, testBody)
	}
	wantJSON := `{"kind":"resume","name":"Mira Example-Halvorsen","location":"Utrecht, NL",` +
		`"email":"mira@example.org","phone":"+00 0 00000000","links":[` +
		`{"label":"github","url":"https://github.com/example"},{"label":"orcid","url":"https://orcid.org/0000"}],` +
		`"position":"Senior Data Scientist","company":"Acme & Co."}`
	if got := compactJSON(t, q.files[DataFile]); got != wantJSON {
		t.Errorf("%s =\n%s\nwant\n%s", DataFile, got, wantJSON)
	}
	for _, f := range []string{"resume.qmd", "typst-template.typ", "partials/style.typ"} {
		if _, ok := q.files[f]; !ok {
			t.Errorf("template file %s not copied", f)
		}
	}
	for _, f := range []string{".hidden", ".git/config"} {
		if _, ok := q.files[f]; ok {
			t.Errorf("dot-entry %s copied", f)
		}
	}
}
