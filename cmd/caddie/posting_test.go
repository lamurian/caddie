package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// `caddie posting save` writes the same posting as the internal/posting
// golden test, merges the review and reports the slug.
func TestPostingSaveCommand(t *testing.T) {
	vars := newWorkspace(t, searchPrefs)
	root := vars["CADDIE_WORKSPACE"]
	review, err := os.ReadFile("../../testdata/reviews/acme.md")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "reviews"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "reviews", "acme.md"), review, 0o600); err != nil {
		t.Fatal(err)
	}
	r := runWith(t, vars, boards, "posting", "save", "https://boards.greenhouse.io/acme/jobs/123")
	if r.code != exitOK || !strings.Contains(r.stdout, `"slug": "senior-data-scientist-at-acme"`) {
		t.Fatalf("exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	got, _ := os.ReadFile(filepath.Join(root, "postings", "senior-data-scientist-at-acme.md"))
	want, _ := os.ReadFile("../../testdata/golden/posting-senior-data-scientist-at-acme.md")
	if string(got) != string(want) {
		t.Errorf("posting differs from the golden file:\n%s", got)
	}
	r = runWith(t, vars, boards, "seen", "list")
	if !strings.Contains(r.stdout, `"status": "saved"`) {
		t.Errorf("not marked saved: %s", r.stdout)
	}
}

func TestPostingSaveErrors(t *testing.T) {
	vars := newWorkspace(t, "")
	r := runCaddie(t, vars, "posting", "save", "https://example.org/gone")
	if r.code != exitExternal || !strings.Contains(r.stderr, "HTTP 404") {
		t.Errorf("404: %+v", r)
	}
}
