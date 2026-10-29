package document

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lamurian/caddie/internal/config"
)

// The real-render tests run the quarto binary (CADDIE_QUARTO_BIN, else
// "quarto" on PATH) and pdftotext, and are skipped when either is missing
// (ADR 0010). The posting's title and company only shape the PDF name.
const (
	realPosition = "Senior Biostatistician"
	realCompany  = "Example Health Alliance"
	realSlug     = "senior-biostatistician-at-example-health-alliance"
	realTimeout  = 3 * time.Minute
)

type realTools struct{ quarto, pdftotext string }

// needRealTools skips the test unless quarto and pdftotext are found.
func needRealTools(t *testing.T) realTools {
	t.Helper()
	bin := os.Getenv("CADDIE_QUARTO_BIN")
	if bin == "" {
		bin = "quarto"
	}
	quarto, err := exec.LookPath(bin)
	if err != nil {
		t.Skipf("real render skipped: %v", err)
	}
	pdftotext, err := exec.LookPath("pdftotext")
	if err != nil {
		t.Skipf("real render skipped: %v", err)
	}
	return realTools{quarto: quarto, pdftotext: pdftotext}
}

// fixtureContact is the fictional contact from the preferences fixture.
func fixtureContact(t *testing.T) config.Contact {
	t.Helper()
	prefs, err := config.LoadPreferences("../../testdata/config/preferences.fixture.yml")
	if err != nil {
		t.Fatal(err)
	}
	if prefs.Contact.Email == "" || prefs.Contact.Phone == "" {
		t.Fatal("preferences fixture has no contact e-mail or phone")
	}
	return prefs.Contact
}

// realRender verifies body as a draft in a fresh workspace and renders it
// with the repository's template, the fixture identity and the fixture
// contact. TMPDIR is a fresh directory outside the repository; it must be
// empty again afterwards. It returns the result, the workspace and the
// PDF's normalised text.
func realRender(t *testing.T, tools realTools, kind Kind, body []byte) (RenderResult, string, string) {
	t.Helper()
	ws, tmp := t.TempDir(), t.TempDir()
	t.Setenv("TMPDIR", tmp)
	dir := map[Kind]string{KindResume: "resumes", KindCoverLetter: "cover-letters"}[kind]
	src := filepath.Join(ws, dir, realSlug+".md")
	writeTestFile(t, src, string(body))
	state := filepath.Join(ws, ".caddie")
	if _, err := RecordVerified(state, dir+"/"+realSlug+".md", body, "2026-10-29"); err != nil {
		t.Fatal(err)
	}
	tpl, err := filepath.Abs(filepath.Join("..", "..", "templates", string(kind)+".qmd"))
	if err != nil {
		t.Fatal(err)
	}
	in := RenderInput{
		Kind: kind, Source: src, StateDir: state,
		HandoffDir: filepath.Join(ws, "handoff"), Template: tpl,
		Identity: fixtureProfile(t).Identity, Contact: fixtureContact(t),
		Position: realPosition, Company: realCompany,
	}
	ctx, cancel := context.WithTimeout(context.Background(), realTimeout)
	defer cancel()
	res, err := Render(ctx, ExecQuarto{Bin: tools.quarto}, in)
	if err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(tmp); len(entries) > 0 {
		t.Errorf("temp render dir left behind: %v", entries)
	}
	return res, ws, pdfText(t, tools.pdftotext, res.PDF)
}

// pdfText extracts the PDF's text with pdftotext -layout and normalises
// it: form feeds become line breaks, trailing blanks and blank lines go.
func pdfText(t *testing.T, bin, pdf string) string {
	t.Helper()
	out, err := exec.Command(bin, "-layout", "-enc", "UTF-8", pdf, "-").Output()
	if err != nil {
		t.Fatalf("pdftotext %s: %v", pdf, err)
	}
	var b strings.Builder
	for _, line := range strings.Split(strings.ReplaceAll(string(out), "\f", "\n"), "\n") {
		if line = strings.TrimRight(line, " \t\r"); line != "" {
			b.WriteString(line + "\n")
		}
	}
	return b.String()
}

// TestRealRenderGolden renders the verified fixtures and compares the
// extracted text with the golden files. Contact details reach the PDF and
// no other file; templates/ is left untouched.
func TestRealRenderGolden(t *testing.T) {
	tools := needRealTools(t)
	templates := snapshot(filepath.Join("..", "..", "templates"))
	contact := fixtureContact(t)
	cases := []struct {
		kind            Kind
		fixture, golden string
	}{
		{KindResume, "resume-traceable.md", "render-resume.txt"},
		{KindCoverLetter, "cover-letter-answers.md", "render-cover-letter.txt"},
	}
	for _, c := range cases {
		t.Run(string(c.kind), func(t *testing.T) {
			body := readFixture(t, filepath.Join("..", "..", "testdata", "documents", c.fixture))
			checkRealGolden(t, tools, c.kind, body, c.golden, contact)
		})
	}
	if after := snapshot(filepath.Join("..", "..", "templates")); !reflect.DeepEqual(templates, after) {
		t.Error("rendering changed templates/ (stray .quarto/ or output?)")
	}
}

func checkRealGolden(t *testing.T, tools realTools, kind Kind, body []byte, golden string, contact config.Contact) {
	t.Helper()
	start := time.Now()
	res, ws, text := realRender(t, tools, kind, body)
	t.Logf("rendered %s in %s", filepath.Base(res.PDF), time.Since(start).Round(time.Millisecond))
	name := PDFName(fixtureProfile(t).Identity.Name, kind, realPosition, realCompany)
	if filepath.Base(res.PDF) != name || filepath.Base(res.Handoff) != name {
		t.Errorf("result %+v, want both named %s", res, name)
	}
	if pdf, handoff := readFixture(t, res.PDF), readFixture(t, res.Handoff); string(pdf) != string(handoff) {
		t.Error("handoff copy differs from the PDF")
	}
	checkGolden(t, filepath.Join("..", "..", "testdata", "golden", golden), []byte(text))
	for _, s := range []string{contact.Email, contact.Phone} {
		if !strings.Contains(text, s) {
			t.Errorf("contact %q missing from the PDF text", s)
		}
	}
	assertContactOnlyInPDFs(t, ws, contact.Email, contact.Phone)
}

// TestRealRenderLiteral: shortcodes, YAML blocks, comments and code cells
// in a verified body print as text; nothing is expanded, read or run
// (without engine: markdown, quarto would try to run the python cell).
func TestRealRenderLiteral(t *testing.T) {
	tools := needRealTools(t)
	t.Setenv("CADDIE_RENDER_PROBE", "probe-value-4711")
	literals := []string{
		"Key {{< env CADDIE_RENDER_PROBE >}} and {{< env HOME >}} end.",
		"Escaped {{{< meta caddie-layout >}}} end.",
		"Include {{< include _caddie.json >}} end.",
		"Open {{< env HOME without a closer.",
		"Escaped open {{{< meta caddie-layout without a closer.",
	}
	body := strings.Join(literals, "\n\n") + "\n\n" +
		"---\nheader-includes: \"injected header\"\n---\n\n" +
		"<!-- c -->---\nheader-includes: \"injected comment\"\n...\n\n" +
		"```{python}\nprint(\"exec\" + \"uted\")\n```\n\nLast line.\n"
	for _, kind := range []Kind{KindResume, KindCoverLetter} {
		t.Run(string(kind), func(t *testing.T) {
			_, _, text := realRender(t, tools, kind, []byte(body))
			checkLiteral(t, text, append(literals, "Last line."))
		})
	}
}

func checkLiteral(t *testing.T, text string, literals []string) {
	t.Helper()
	flat := strings.Join(strings.Fields(text), " ")
	for _, want := range literals {
		if !strings.Contains(flat, want) {
			t.Errorf("literal %q missing from:\n%s", want, text)
		}
	}
	for _, leak := range []string{"probe-value-4711", "executed", `"kind"`} {
		if strings.Contains(text, leak) {
			t.Errorf("%q expanded into the PDF:\n%s", leak, text)
		}
	}
	if home := os.Getenv("HOME"); home != "" && strings.Contains(text, home) {
		t.Errorf("$HOME (%s) expanded into the PDF:\n%s", home, text)
	}
	for _, s := range []string{"injected header", "injected comment"} {
		if n := strings.Count(text, s); n != 1 {
			t.Errorf("%q appears %d times, want once as body text:\n%s", s, n, text)
		}
	}
	if first, _, _ := strings.Cut(strings.TrimSpace(text), "\n"); first != fixtureProfile(t).Identity.Name {
		t.Errorf("first line %q, want the name: something was put before the header", first)
	}
}
