package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/workspace"
)

// moveWorkspace is renderWorkspace plus both PDFs, a handoff copy and a
// trace screenshot.
func moveWorkspace(t *testing.T, status string) (map[string]string, string) {
	t.Helper()
	vars, root := renderWorkspace(t, status)
	writeFile(t, filepath.Join(root, "resumes", resumePDF), "%PDF-r")
	writeFile(t, filepath.Join(root, "cover-letters", letterPDF), "%PDF-l")
	writeFile(t, filepath.Join(root, "handoff", resumePDF), "%PDF-r")
	writeFile(t, filepath.Join(root, ".caddie", "trace", "01-filled.png"), "png")
	return vars, root
}

func TestMoveApplied(t *testing.T) {
	vars, root := moveWorkspace(t, "rendered")
	r := runCaddie(t, vars, "move", "--json", acmeSlug, "applied")
	if r.code != exitOK || r.stderr != "" {
		t.Fatalf("exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	var out workspace.MoveResult
	if err := json.Unmarshal([]byte(r.stdout), &out); err != nil {
		t.Fatal(err)
	}
	dir := "applied/" + acmeSlug + "/"
	wantMoved := []string{dir + "resume.md", dir + "cover-letter.md", dir + resumePDF,
		dir + letterPDF, dir + "trace/01-filled.png"}
	if out.Status != "applied" || out.Posting != "applied/"+acmeSlug+".md" ||
		out.AppliedAt != "2026-10-28" || out.FollowUpAfter != "2026-11-11" ||
		!slices.Equal(out.Moved, wantMoved) || len(out.Skipped) != 0 {
		t.Errorf("result %+v", out)
	}
	checkMovedFiles(t, root, append(wantMoved, out.Posting))
	r = runCaddie(t, vars, "move", acmeSlug, "applied")
	if r.code != exitIllegalState {
		t.Errorf("second move: exit %d: %s", r.code, r.stderr)
	}
}

// checkMovedFiles asserts the moved files exist, the sources, handoff copy
// and trace files are gone and the trace directory stays.
func checkMovedFiles(t *testing.T, root string, moved []string) {
	t.Helper()
	for _, p := range moved {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Errorf("not moved: %v", err)
		}
	}
	for _, p := range []string{postingFile(root), filepath.Join(root, "handoff", resumePDF),
		filepath.Join(root, ".caddie", "trace", "01-filled.png")} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s still there: %v", p, err)
		}
	}
	if fi, err := os.Stat(filepath.Join(root, ".caddie", "trace")); err != nil || !fi.IsDir() {
		t.Errorf("trace dir gone: %v", err)
	}
}

func TestMoveManualNeedsReason(t *testing.T) {
	vars, root := moveWorkspace(t, "saved")
	r := runCaddie(t, vars, "move", acmeSlug, "manual")
	if r.code != exitUsage || !strings.Contains(r.stderr, "reason") {
		t.Errorf("no reason: exit %d: %s", r.code, r.stderr)
	}
	r = runCaddie(t, vars, "move", acmeSlug, "manual", "--reason", "Workday login wall")
	if r.code != exitOK || !strings.Contains(r.stdout, `"manual_reason": "Workday login wall"`) {
		t.Fatalf("exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	if _, err := os.Stat(filepath.Join(root, "manual", acmeSlug+".md")); err != nil {
		t.Error(err)
	}
}

func TestMoveRefusals(t *testing.T) {
	cases := []struct {
		name, status string
		prep         func(root string)
		args         []string
		code         int
		want         string
	}{
		{"saved to applied", "saved", nil, []string{acmeSlug, "applied"}, exitIllegalState, "saved"},
		{"dest exists", "rendered", func(root string) {
			writeFile(t, filepath.Join(root, "applied", acmeSlug, "x"), "")
		}, []string{acmeSlug, "applied"}, exitIllegalState, "already exists"},
		{"unknown slug", "rendered", nil, []string{"nope", "applied"}, exitNotFound, "no posting postings/nope.md"},
		{"bad target", "rendered", nil, []string{acmeSlug, "postings"}, exitUsage, "applied or manual"},
		{"bad slug", "rendered", nil, []string{"../x", "applied"}, exitUsage, "slug"},
		{"reason on applied", "rendered", nil, []string{acmeSlug, "applied", "--reason", "x"}, exitUsage, "reason"},
		{"one arg", "rendered", nil, []string{acmeSlug}, exitUsage, "want `move"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vars, root := moveWorkspace(t, c.status)
			if c.prep != nil {
				c.prep(root)
			}
			r := runCaddie(t, vars, append([]string{"move"}, c.args...)...)
			if r.code != c.code || !strings.Contains(r.stderr, c.want) || r.stdout != "" {
				t.Errorf("exit %d, want %d: %q %q", r.code, c.code, r.stdout, r.stderr)
			}
			if _, err := os.Stat(postingFile(root)); err != nil {
				t.Errorf("posting moved: %v", err)
			}
		})
	}
}

func TestPostingSteps(t *testing.T) {
	vars, root := moveWorkspace(t, "saved")
	r := runCaddie(t, vars, "posting", "steps", acmeSlug, "Create a Workday account", "Upload the resume")
	var out struct {
		Path  string `json:"path"`
		Steps []struct {
			Text string `json:"text"`
			Done bool   `json:"done"`
		} `json:"steps"`
		Added int `json:"added"`
	}
	if err := json.Unmarshal([]byte(r.stdout), &out); err != nil || r.code != exitOK {
		t.Fatalf("exit %d: %v: %s%s", r.code, err, r.stdout, r.stderr)
	}
	if out.Path != "postings/"+acmeSlug+".md" || out.Added != 2 || len(out.Steps) != 2 {
		t.Errorf("first run %+v", out)
	}
	data, _ := os.ReadFile(postingFile(root))
	if !strings.Contains(string(data), "## Manual steps\n\n- [ ] Create a Workday account\n- [ ] Upload the resume\n") {
		t.Errorf("posting:\n%s", data)
	}
	r = runCaddie(t, vars, "posting", "steps", acmeSlug, "upload the RESUME")
	if r.code != exitOK || !strings.Contains(r.stdout, `"added": 0`) {
		t.Errorf("second run: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	if r := runCaddie(t, vars, "posting", "steps", acmeSlug); r.code != exitUsage {
		t.Errorf("no steps: exit %d: %s", r.code, r.stderr)
	}
	if r := runCaddie(t, vars, "posting", "steps", acmeSlug, " "); r.code != exitUsage {
		t.Errorf("blank step: exit %d: %s", r.code, r.stderr)
	}
	if r := runCaddie(t, vars, "posting", "steps", "nope", "x"); r.code != exitNotFound {
		t.Errorf("unknown: exit %d: %s", r.code, r.stderr)
	}
}

func TestPostingStepsManualAndApplied(t *testing.T) {
	vars, root := moveWorkspace(t, "rendered")
	r := runCaddie(t, vars, "move", acmeSlug, "manual", "--reason", "login wall")
	if r.code != exitOK {
		t.Fatalf("move: exit %d: %s", r.code, r.stderr)
	}
	r = runCaddie(t, vars, "posting", "steps", acmeSlug, "Apply by email")
	if r.code != exitOK || !strings.Contains(r.stdout, `"path": "manual/`+acmeSlug+`.md"`) {
		t.Errorf("manual: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	if err := os.Rename(filepath.Join(root, "manual", acmeSlug+".md"),
		filepath.Join(root, "applied", acmeSlug+".md")); err != nil {
		t.Fatal(err)
	}
	r = runCaddie(t, vars, "posting", "steps", acmeSlug, "x")
	if r.code != exitIllegalState || !strings.Contains(r.stderr, "applied/") {
		t.Errorf("applied: exit %d: %s", r.code, r.stderr)
	}
	if r := runCaddie(t, vars, "posting", "bogus"); r.code != exitUsage ||
		!strings.Contains(r.stderr, "posting steps <slug>") {
		t.Errorf("bogus: exit %d: %s", r.code, r.stderr)
	}
}
