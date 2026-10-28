package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
)

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Senior Data Scientist":      "senior-data-scientist",
		"  Société Générale (Paris)": "societe-generale-paris",
		"R&D / ML -- Lead!":          "r-d-ml-lead",
		"Ünïcödé 2026":               "unicode-2026",
		"日本":                         "",
		strings.Repeat("abc ", 40):   strings.TrimSuffix(strings.Repeat("abc-", 20), "-"),
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
	if got := PostingSlug("Senior Data Scientist", "Acme"); got != "senior-data-scientist-at-acme" {
		t.Errorf("PostingSlug = %q", got)
	}
}

func TestLayout(t *testing.T) {
	root := t.TempDir()
	cfg := config.Config{Workspace: root, HandoffDir: "handoff", StateDir: ".caddie",
		Preferences: "preferences.yml", Answers: "/abs/answers.yml"}
	l := New(cfg)
	if l.Handoff != filepath.Join(root, "handoff") || l.Answers != "/abs/answers.yml" {
		t.Errorf("paths not resolved: %+v", l)
	}
	if n := len(l.Missing()); n != len(l.Required()) {
		t.Errorf("empty workspace: %d missing, want %d", n, len(l.Required()))
	}
	for _, d := range l.Required() {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if m := l.Missing(); len(m) != 0 {
		t.Errorf("complete workspace reports missing: %v", m)
	}
}

func TestParsePosting(t *testing.T) {
	name := "../../testdata/postings/senior-data-scientist-at-acme.md"
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	p, body, err := ParsePosting(name, data)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != StatusSaved || p.Deadline != "2026-11-14" || p.SalaryMax != 95000 || p.AppliedAt != "" {
		t.Errorf("unexpected posting: %+v", p)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(body)), "## Description") {
		t.Errorf("body = %q", body)
	}
}

func TestParsePostingErrors(t *testing.T) {
	cases := map[string]string{
		"no frontmatter\n":                         "p.md:1: missing --- frontmatter",
		"---\nslug: x\n":                           "missing --- frontmatter",
		"---\nslug: x\nstatus: done\n---\n":        "unknown posting status",
		"---\nstatus: saved\n---\n":                "slug is empty",
		"---\nslug: x\nstatus: saved\nfoo: 1\n---": "line 4",
	}
	for in, want := range cases {
		_, _, err := ParsePosting("p.md", []byte(in))
		if !errors.Is(err, ErrBadPosting) || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: got %v, want %q", in, err, want)
		}
	}
}
