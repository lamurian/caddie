package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lamurian/caddie/internal/document"
)

// recordedPath is the path the verified record for content names.
func recordedPath(t *testing.T, state string, content []byte) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(state, document.VerifiedDir, document.Hash(content)))
	if err != nil {
		t.Fatal(err)
	}
	var rec struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatal(err)
	}
	return rec.Path
}

// However the draft is named on the command line, the verified record
// names it relative to the workspace root; a draft outside the workspace
// is named by its absolute path.
func TestVerifyRecordPath(t *testing.T) {
	vars, path := draftWorkspace(t, "resume-traceable.md", "resumes", acmeSlug)
	root := vars["CADDIE_WORKSPACE"]
	outside := filepath.Join(t.TempDir(), "drafts", acmeSlug+".md")
	copyFile(t, path, outside)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	draft := "resumes/" + acmeSlug + ".md"
	sub := map[string]string{"CADDIE_WORKSPACE": filepath.Base(root)}
	cases := []struct {
		name, cwd string
		vars      map[string]string
		arg, want string
	}{
		{"dot-relative", root, vars, "./" + draft, draft},
		{"relative", root, vars, draft, draft},
		{"absolute", t.TempDir(), vars, path, draft},
		{"workspace subdir", filepath.Dir(root), sub, filepath.Join(filepath.Base(root), draft), draft},
		{"outside", root, vars, outside, outside},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Chdir(c.cwd)
			if r := runCaddie(t, c.vars, "verify", c.arg); r.code != exitOK {
				t.Fatalf("exit %d: %s%s", r.code, r.stdout, r.stderr)
			}
			if got := recordedPath(t, filepath.Join(root, ".caddie"), content); got != c.want {
				t.Errorf("verify %s from %s: record path %q, want %q", c.arg, c.cwd, got, c.want)
			}
		})
	}
}
