package workspace

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// Every refusal happens before anything moves: the tree is byte-identical.
func TestMoveRefusals(t *testing.T) {
	cases := []struct {
		name   string
		status Status
		setup  func(t *testing.T, l Layout, o *MoveOptions)
		err    error
	}{
		{"dest posting exists", StatusRendered, func(t *testing.T, l Layout, _ *MoveOptions) {
			writeT(t, filepath.Join(l.Dir(DirApplied), mvSlug+".md"), "old", 0o644)
		}, ErrDestExists},
		{"dest dir exists", StatusRendered, func(t *testing.T, l Layout, _ *MoveOptions) {
			mkdirT(t, filepath.Join(l.Dir(DirApplied), mvSlug))
		}, ErrDestExists},
		{"applied from saved", StatusSaved, nil, ErrIllegalTransition},
		{"manual from applied", StatusApplied, func(_ *testing.T, _ Layout, o *MoveOptions) {
			o.To, o.Reason = StatusManual, "late"
		}, ErrIllegalTransition},
		{"already archived", StatusRendered, func(t *testing.T, l Layout, _ *MoveOptions) {
			src := filepath.Join(l.Dir(DirPostings), mvSlug+".md")
			if err := os.Rename(src, filepath.Join(l.Dir(DirManual), mvSlug+".md")); err != nil {
				t.Fatal(err)
			}
		}, ErrIllegalTransition},
		{"not found", StatusRendered, func(_ *testing.T, _ Layout, o *MoveOptions) { o.Slug = "nope" }, os.ErrNotExist},
		{"target rendered", StatusTailored, func(_ *testing.T, _ Layout, o *MoveOptions) { o.To = StatusRendered }, ErrIllegalTransition},
		{"manual without reason", StatusRendered, func(_ *testing.T, _ Layout, o *MoveOptions) { o.To, o.Reason = StatusManual, " " }, ErrBadMove},
		{"two-line reason", StatusRendered, func(_ *testing.T, _ Layout, o *MoveOptions) { o.To, o.Reason = StatusManual, "a\nb" }, ErrBadMove},
		{"applied with reason", StatusRendered, func(_ *testing.T, _ Layout, o *MoveOptions) { o.Reason = "x" }, ErrBadMove},
		{"bad today", StatusRendered, func(_ *testing.T, _ Layout, o *MoveOptions) { o.Today = "2026-13-01" }, ErrBadMove},
		{"negative days", StatusRendered, func(_ *testing.T, _ Layout, o *MoveOptions) { o.FollowUpDays = -1 }, ErrBadMove},
		{"slug escapes", StatusRendered, func(_ *testing.T, _ Layout, o *MoveOptions) { o.Slug = "../postings/x" }, ErrBadMove},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l, o, _ := moveWS(t, c.status)
			if c.setup != nil {
				c.setup(t, l, &o)
			}
			before := tree(t, l.Root)
			if _, err := Move(l, o); !errors.Is(err, c.err) {
				t.Errorf("err %v, want %v", err, c.err)
			}
			sameTree(t, before, tree(t, l.Root))
		})
	}
}

// An injected failure at any step leaves everything where it started.
func TestMoveRollback(t *testing.T) {
	cases := []struct {
		name   string
		inject func(t *testing.T)
	}{
		{"first rename", func(t *testing.T) { failOn(t, 1) }},
		{"third rename", func(t *testing.T) { failOn(t, 3) }},
		{"last trace rename", func(t *testing.T) { failOn(t, 6) }},
		{"posting write", func(t *testing.T) {
			writePosting = func(string, []byte, os.FileMode) error { return fs.ErrPermission }
			t.Cleanup(func() { writePosting = writeNew })
		}},
		{"source remove", func(t *testing.T) {
			removeFile = func(string) error { return fs.ErrPermission }
			t.Cleanup(func() { removeFile = os.Remove })
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l, o, _ := moveWS(t, StatusRendered)
			before := tree(t, l.Root)
			c.inject(t)
			_, err := Move(l, o)
			if !errors.Is(err, fs.ErrPermission) || strings.Contains(err.Error(), "could not undo") {
				t.Errorf("err %v", err)
			}
			sameTree(t, before, tree(t, l.Root))
		})
	}
}

// When an undo also fails the error names the step left behind.
func TestMoveUndoFailureNamed(t *testing.T) {
	l, o, _ := moveWS(t, StatusRendered)
	failOn(t, 3, 5) // 3: the resume PDF fails; 4, 5: undo cover-letter.md, resume.md (fails)
	_, err := Move(l, o)
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("err %v", err)
	}
	msg := err.Error()
	if !strings.Contains(msg, "could not undo") || !strings.Contains(msg, "resume.md") ||
		!strings.Contains(msg, filepath.Join(l.Dir(DirApplied), mvSlug)) {
		t.Errorf("error does not name the stuck steps: %s", msg)
	}
	if _, serr := os.Stat(filepath.Join(l.Dir(DirPostings), mvSlug+".md")); serr != nil {
		t.Errorf("source posting gone: %v", serr)
	}
}

// Across devices a rename falls back to copy and remove, with rollback too.
func TestMoveCrossDevice(t *testing.T) {
	exdev := func(t *testing.T) {
		rename = func(from, to string) error {
			return &os.LinkError{Op: "rename", Old: from, New: to, Err: syscall.EXDEV}
		}
		t.Cleanup(func() { rename = os.Rename })
	}
	l, o, _ := moveWS(t, StatusRendered)
	exdev(t)
	res, err := Move(l, o)
	if err != nil || len(res.Moved) != 6 {
		t.Fatalf("got %+v, %v", res, err)
	}
	got := tree(t, l.Root)
	if got[filepath.Join("applied", mvSlug, mvResume)] != "-rw-r--r-- %PDF resume" {
		t.Errorf("copied pdf: %q", got[filepath.Join("applied", mvSlug, mvResume)])
	}
	if _, ok := got[filepath.Join("resumes", mvResume)]; ok {
		t.Error("source pdf left behind after copy")
	}

	l, o, _ = moveWS(t, StatusRendered)
	before := tree(t, l.Root)
	writePosting = func(string, []byte, os.FileMode) error { return fs.ErrPermission }
	t.Cleanup(func() { writePosting = writeNew })
	if _, err := Move(l, o); !errors.Is(err, fs.ErrPermission) {
		t.Errorf("err %v", err)
	}
	sameTree(t, before, tree(t, l.Root))
}
