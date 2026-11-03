package workspace

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestStatusPath(t *testing.T) {
	cases := []struct {
		from, to Status
		want     []Status
		err      error
	}{
		{StatusSaved, StatusRendered, []Status{StatusTailored, StatusRendered}, nil},
		{StatusSaved, StatusTailored, []Status{StatusTailored}, nil},
		{StatusTailored, StatusRendered, []Status{StatusRendered}, nil},
		{StatusSaved, StatusApplied, []Status{StatusTailored, StatusRendered, StatusApplied}, nil},
		{StatusTailored, StatusManual, []Status{StatusManual}, nil},
		{StatusRendered, StatusRendered, nil, nil},
		{StatusSaved, StatusSaved, nil, nil},
		{StatusApplied, StatusRendered, nil, ErrIllegalTransition},
		{StatusManual, StatusRendered, nil, ErrIllegalTransition},
		{StatusRendered, StatusTailored, nil, ErrIllegalTransition},
		{StatusApplied, StatusManual, nil, ErrIllegalTransition},
		{"drafted", StatusRendered, nil, ErrUnknownStatus},
		{StatusSaved, "submitted", nil, ErrUnknownStatus},
	}
	for _, c := range cases {
		got, err := StatusPath(c.from, c.to)
		if !errors.Is(err, c.err) || (c.err == nil && err != nil) {
			t.Errorf("%s → %s: err %v, want %v", c.from, c.to, err, c.err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s → %s: steps %v, want %v", c.from, c.to, got, c.want)
		}
	}
}

// body is appended to the fixture's own body: a later "---" rule, a
// reviews section, trailing spaces and no final newline must all survive.
const body = "\n---\n\n## Reviews\n\nGood team.  \n\tIndented line"

// posting copies the Acme fixture into a temp dir with the given status
// and mode, and returns its path and the body after the frontmatter.
func posting(t *testing.T, status Status, mode os.FileMode) (string, []byte) {
	t.Helper()
	data, err := os.ReadFile("../../testdata/postings/senior-data-scientist-at-acme.md")
	if err != nil {
		t.Fatal(err)
	}
	p, b, err := ParsePosting("fixture", data)
	if err != nil {
		t.Fatal(err)
	}
	p.Status = status
	b = append(b, body...)
	path := filepath.Join(t.TempDir(), "senior-data-scientist-at-acme.md")
	if err := os.WriteFile(path, append(p.Frontmatter(), b...), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
	old := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	return path, b
}

func TestAdvanceStatusRewrites(t *testing.T) {
	for _, from := range []Status{StatusSaved, StatusTailored} {
		t.Run(string(from), func(t *testing.T) {
			path, wantBody := posting(t, from, 0o600)
			p, err := AdvanceStatus(path, StatusRendered)
			if err != nil {
				t.Fatal(err)
			}
			if p.Status != StatusRendered || p.Slug != "senior-data-scientist-at-acme" {
				t.Errorf("returned %+v", p)
			}
			checkRewritten(t, path, p, wantBody)
		})
	}
}

// checkRewritten asserts the file at path holds p's frontmatter, the
// original body, mode 0600 and no leftover temp file.
func checkRewritten(t *testing.T, path string, p Posting, wantBody []byte) {
	t.Helper()
	data, fi := snapshot(t, path)
	back, gotBody, err := ParsePosting(path, data)
	if err != nil || back != p {
		t.Errorf("on disk %+v (%v), want %+v", back, err, p)
	}
	if !bytes.Equal(gotBody, wantBody) {
		t.Errorf("body changed:\n%q\nwant\n%q", gotBody, wantBody)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("mode %v, want 0600", fi.Mode().Perm())
	}
	if left, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".*tmp*")); len(left) > 0 {
		t.Errorf("temp files left: %v", left)
	}
}

// A no-op or a refused change leaves the file's bytes and mtime alone.
func TestAdvanceStatusLeavesFile(t *testing.T) {
	cases := []struct {
		from Status
		err  error
	}{
		{StatusRendered, nil},
		{StatusApplied, ErrIllegalTransition},
		{StatusManual, ErrIllegalTransition},
	}
	for _, c := range cases {
		path, _ := posting(t, c.from, 0o644)
		before, beforeInfo := snapshot(t, path)
		p, err := AdvanceStatus(path, StatusRendered)
		if !errors.Is(err, c.err) || (c.err == nil && err != nil) {
			t.Errorf("%s: err %v, want %v", c.from, err, c.err)
		}
		if c.err == nil && p.Status != StatusRendered {
			t.Errorf("%s: returned status %s", c.from, p.Status)
		}
		after, afterInfo := snapshot(t, path)
		if !bytes.Equal(before, after) || !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
			t.Errorf("%s: file was rewritten", c.from)
		}
	}
}

func TestAdvanceStatusErrors(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.md")
	if err := os.WriteFile(bad, []byte("---\nslug: x\nstatus: done\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before, _ := snapshot(t, bad)
	if _, err := AdvanceStatus(bad, StatusRendered); !errors.Is(err, ErrBadPosting) {
		t.Errorf("malformed: got %v, want ErrBadPosting", err)
	}
	if after, _ := snapshot(t, bad); !bytes.Equal(before, after) {
		t.Error("malformed posting was rewritten")
	}
	if _, err := AdvanceStatus(filepath.Join(dir, "none.md"), StatusRendered); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("missing: got %v, want os.ErrNotExist", err)
	}
	path, _ := posting(t, StatusSaved, 0o644)
	_, err := AdvanceStatus(path, "submitted")
	if !errors.Is(err, ErrUnknownStatus) || !strings.Contains(err.Error(), path) {
		t.Errorf("unknown target: got %v", err)
	}
}

func snapshot(t *testing.T, path string) ([]byte, os.FileInfo) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return data, fi
}
