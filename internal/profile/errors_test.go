package profile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const goodAbout = "---\nname: Test Person\n---\n## About\nFictional.\n"

// writeHistory creates a history dir with a valid about.md plus files.
func writeHistory(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	all := map[string]string{"about.md": goodAbout}
	for k, v := range files {
		all[k] = v
	}
	for name, content := range all {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestSectionErrorsNameFileAndLine(t *testing.T) {
	cases := []struct{ name, content, want string }{
		{"no heading", "### Title\nOrg\nJan 2020 – Present\n", "employment.md:1: missing"},
		{"text before entry", "## Employment\nstray text\n", "employment.md:2: text before"},
		{"two headings", "## A\n### T\nOrg\n2020 – 2021\n## B\n", "employment.md:5: second ##"},
		{"deep heading", "## A\n#### T\n", "employment.md:2: unexpected heading"},
		{"missing dates", "## A\n\n### Analyst\nOrg\n- bullet\n", "employment.md:3: entry \"Analyst\""},
		{"too many lines", "## A\n### T\nOrg\nCity\nExtra\n2020 – 2021\n", "employment.md:2:"},
		{"bad range", "## A\n### T\nOrg\nCity\nJan 2020 to Mar 2021\n", "employment.md:5: bad date range"},
		{"bad month", "## A\n### T\nOrg\nFoo 2020 – Present\n", "employment.md:4: bad month"},
		{"reversed", "## A\n### T\nOrg\nMar 2021 – Jan 2020\n", "employment.md:4: date range"},
		{"text after bullets", "## A\n### T\nOrg\n2020 – 2021\n- b\nloose\n", "employment.md:6: expected a bullet"},
	}
	for _, c := range cases {
		dir := writeHistory(t, map[string]string{"employment.md": c.content})
		_, err := Load(dir)
		if !errors.Is(err, ErrMalformed) || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: got %v, want %q", c.name, err, c.want)
		}
	}
}

func TestWorkErrors(t *testing.T) {
	cases := []struct{ content, want string }{
		{"## P\n### Paper\nVenue\n2020\n", "publication.md:2: \"Paper\" needs venue"},
		{"## P\n### Paper\nVenue\nA, B\nSoon\n", "publication.md:5: bad date"},
	}
	for _, c := range cases {
		dir := writeHistory(t, map[string]string{"publication.md": c.content})
		if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("got %v, want %q", err, c.want)
		}
	}
}

func TestAboutErrors(t *testing.T) {
	cases := []struct{ about, want string }{
		{"name: X\n", "about.md:1: about.md must start with ---"},
		{"---\nname: X\n", "about.md:1: frontmatter has no closing"},
		{"---\nname: X\nnationalty: NO\n---\n## About\nx\n", "about.md: frontmatter: yaml: unmarshal errors:\n  line 3"},
		{"---\nname: [\n---\n## About\nx\n", "about.md: frontmatter: yaml: line"},
		{"---\nlocation: X\n---\n## About\nx\n", "about.md:2: frontmatter: name is required"},
		{"---\nname: X\n---\n\n## Bio\nx\n", "about.md:5: expected \"## About\""},
		{"---\nname: X\n---\n## About\n\n", "about.md:4: the ## About section is empty"},
		{"---\nname: X\n---\n", "missing \"## About\""},
	}
	for _, c := range cases {
		dir := writeHistory(t, map[string]string{"about.md": c.about})
		_, err := Load(dir)
		if !errors.Is(err, ErrMalformed) || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%q: got %v, want %q", c.about, err, c.want)
		}
	}
}

func TestMissingAbout(t *testing.T) {
	_, err := Load(t.TempDir())
	if !errors.Is(err, ErrMalformed) || !strings.Contains(err.Error(), "about.md: missing") {
		t.Errorf("got %v", err)
	}
}

func TestReadmeAndSubdirsSkipped(t *testing.T) {
	dir := writeHistory(t, map[string]string{"README.md": "# not a section\n", "notes.txt": "x"})
	if err := os.Mkdir(filepath.Join(dir, "drafts.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	p, err := Load(dir)
	if err != nil || len(p.Sections) != 0 {
		t.Errorf("got %+v, %v", p.Sections, err)
	}
}
