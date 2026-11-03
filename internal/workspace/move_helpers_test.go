package workspace

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

const (
	mvSlug    = "senior-data-scientist-at-acme"
	mvResume  = "Resume-Jane-Doe-Acme.pdf"
	mvLetter  = "Cover-Letter-Jane-Doe-Acme.pdf"
	mvToday   = "2026-10-28"
	mvBodyEnd = "\n---\n\n## Manual steps\n\n- [x] Done  \n\tno final newline"
)

// moveWS builds a workspace in a temp dir holding the Acme posting with the
// given status, every artefact, two trace files and a stocked handoff/.
// It returns the layout, the options for a move and the posting body.
func moveWS(t *testing.T, status Status) (Layout, MoveOptions, []byte) {
	t.Helper()
	root := t.TempDir()
	l := Layout{Root: root, Handoff: filepath.Join(root, "handoff"), State: filepath.Join(root, ".caddie")}
	trace := filepath.Join(l.State, "trace")
	for _, d := range []string{DirPostings, DirApplied, DirManual, DirResumes, DirCoverLetters,
		"handoff/sub", ".caddie/trace"} {
		mkdirT(t, filepath.Join(root, d))
	}
	data, err := os.ReadFile("../../testdata/postings/" + mvSlug + ".md")
	if err != nil {
		t.Fatal(err)
	}
	p, body, err := ParsePosting("fixture", data)
	if err != nil {
		t.Fatal(err)
	}
	p.Status = status
	body = append(body, mvBodyEnd...)
	writeT(t, filepath.Join(l.Dir(DirPostings), mvSlug+".md"), string(append(p.Frontmatter(), body...)), 0o600)
	files := map[string]string{
		"resumes/" + mvSlug + ".md":       "# resume",
		"resumes/" + mvResume:             "%PDF resume",
		"cover-letters/" + mvSlug + ".md": "# letter",
		"cover-letters/" + mvLetter:       "%PDF letter",
		".caddie/trace/01-filled.png":     "png1",
		".caddie/trace/02-confirmed.png":  "png2",
		"handoff/.gitignore":              "*\n!.gitignore\n!README.md\n",
		"handoff/README.md":               "# handoff/",
		"handoff/" + mvResume:             "%PDF resume",
		"handoff/sub/keep.txt":            "kept",
	}
	for name, content := range files {
		writeT(t, filepath.Join(root, name), content, 0o644)
	}
	o := MoveOptions{Slug: mvSlug, To: StatusApplied, Today: mvToday, FollowUpDays: 7,
		ResumePDF: mvResume, LetterPDF: mvLetter, TraceDir: trace}
	return l, o, body
}

func mkdirT(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeT(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

// tree maps every path under root to its content and mode ("dir" for
// directories), so two snapshots are equal only if the tree is identical.
func tree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		r, _ := filepath.Rel(root, path)
		if d.IsDir() {
			out[r] = "dir"
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		out[r] = fi.Mode().String() + " " + string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// sameTree reports the first difference between two snapshots.
func sameTree(t *testing.T, before, after map[string]string) {
	t.Helper()
	for k, v := range before {
		if got, ok := after[k]; !ok || got != v {
			t.Errorf("%s changed or missing: %q → %q", k, v, got)
		}
	}
	for k := range after {
		if _, ok := before[k]; !ok {
			t.Errorf("%s appeared", k)
		}
	}
}

// failOn makes the n-th call of the injectable rename fail and restores it
// when the test ends.
func failOn(t *testing.T, n ...int) {
	t.Helper()
	calls := 0
	fail := map[int]bool{}
	for _, i := range n {
		fail[i] = true
	}
	rename = func(from, to string) error {
		calls++
		if fail[calls] {
			return &os.LinkError{Op: "rename", Old: from, New: to, Err: fs.ErrPermission}
		}
		return os.Rename(from, to)
	}
	t.Cleanup(func() { rename = os.Rename })
}
