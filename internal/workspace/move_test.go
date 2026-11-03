package workspace

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMoveApplied(t *testing.T) {
	l, o, wantBody := moveWS(t, StatusRendered)
	res, err := Move(l, o)
	if err != nil {
		t.Fatal(err)
	}
	dir := "applied/" + mvSlug + "/"
	want := MoveResult{
		Slug: mvSlug, Status: StatusApplied, Posting: "applied/" + mvSlug + ".md",
		Moved: []string{dir + "resume.md", dir + "cover-letter.md", dir + mvResume, dir + mvLetter,
			dir + "trace/01-filled.png", dir + "trace/02-confirmed.png"},
		Skipped: []string{}, AppliedAt: mvToday, FollowUpAfter: "2026-11-04",
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("result\n%+v\nwant\n%+v", res, want)
	}
	p := checkArchived(t, l, res.Posting, wantBody)
	if p.Status != StatusApplied || p.AppliedAt != mvToday || p.FollowUpAfter != "2026-11-04" || p.ManualReason != "" {
		t.Errorf("frontmatter %+v", p)
	}
	got := tree(t, l.Root)
	for _, f := range res.Moved {
		if _, ok := got[filepath.FromSlash(f)]; !ok {
			t.Errorf("%s not moved", f)
		}
	}
	for _, gone := range []string{"postings/" + mvSlug + ".md", "resumes/" + mvSlug + ".md",
		"resumes/" + mvResume, "cover-letters/" + mvSlug + ".md", "cover-letters/" + mvLetter,
		".caddie/trace/01-filled.png", "handoff/" + mvResume} {
		if _, ok := got[filepath.FromSlash(gone)]; ok {
			t.Errorf("%s still present", gone)
		}
	}
	for _, kept := range []string{"handoff/.gitignore", "handoff/sub/keep.txt", ".caddie/trace"} {
		if _, ok := got[filepath.FromSlash(kept)]; !ok {
			t.Errorf("%s removed", kept)
		}
	}
	if got[filepath.FromSlash(dir+"resume.md")] != "-rw-r--r-- # resume" {
		t.Errorf("resume.md content %q", got[filepath.FromSlash(dir+"resume.md")])
	}
}

// checkArchived parses the archived posting, checks its body is the
// original byte for byte, its mode 0600 and no temp file is left.
func checkArchived(t *testing.T, l Layout, posting string, wantBody []byte) Posting {
	t.Helper()
	path := filepath.Join(l.Root, filepath.FromSlash(posting))
	data, fi := snapshot(t, path)
	p, body, err := ParsePosting(path, data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, wantBody) {
		t.Errorf("body changed:\n%q\nwant\n%q", body, wantBody)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("mode %v, want 0600", fi.Mode().Perm())
	}
	if left, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".*tmp*")); len(left) > 0 {
		t.Errorf("temp files left: %v", left)
	}
	return p
}

func TestMoveFollowUpMath(t *testing.T) {
	cases := []struct {
		today string
		days  int
		want  string
	}{
		{"2026-01-30", 3, "2026-02-02"},
		{"2026-12-29", 7, "2027-01-05"},
		{"2028-02-27", 2, "2028-02-29"},
		{"2026-11-03", 0, "2026-11-03"},
	}
	for _, c := range cases {
		l, o, _ := moveWS(t, StatusRendered)
		o.Today, o.FollowUpDays = c.today, c.days
		res, err := Move(l, o)
		if err != nil || res.FollowUpAfter != c.want || res.AppliedAt != c.today {
			t.Errorf("%s + %d: %+v (%v), want %s", c.today, c.days, res, err, c.want)
		}
	}
}

// Manual keeps the dates, records the reason and tolerates every optional
// artefact being absent.
func TestMoveManualMissingArtefacts(t *testing.T) {
	l, o, wantBody := moveWS(t, StatusTailored)
	for _, f := range []string{"resumes/" + mvResume, "cover-letters/" + mvSlug + ".md",
		"cover-letters/" + mvLetter, ".caddie/trace/01-filled.png", ".caddie/trace/02-confirmed.png"} {
		if err := os.Remove(filepath.Join(l.Root, f)); err != nil {
			t.Fatal(err)
		}
	}
	o.To, o.Reason, o.Today, o.FollowUpDays = StatusManual, "  login wall on Workday ", "", 0
	res, err := Move(l, o)
	if err != nil {
		t.Fatal(err)
	}
	want := MoveResult{
		Slug: mvSlug, Status: StatusManual, Posting: "manual/" + mvSlug + ".md",
		Moved: []string{"manual/" + mvSlug + "/resume.md"},
		Skipped: []string{"cover-letters/" + mvSlug + ".md", "resumes/" + mvResume,
			"cover-letters/" + mvLetter, ".caddie/trace"},
		ManualReason: "login wall on Workday",
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("result\n%+v\nwant\n%+v", res, want)
	}
	p := checkArchived(t, l, res.Posting, wantBody)
	if p.Status != StatusManual || p.ManualReason != want.ManualReason || p.AppliedAt != "" ||
		p.SavedAt != "2026-10-28" || p.Deadline != "2026-11-14" {
		t.Errorf("frontmatter %+v", p)
	}
	if _, err := os.Stat(filepath.Join(l.Root, "manual", mvSlug, "trace")); !os.IsNotExist(err) {
		t.Errorf("empty trace dir created: %v", err)
	}
}

// No PDF names and no trace dir: nothing to look for, nothing skipped.
func TestMoveNoOptionalNames(t *testing.T) {
	l, o, _ := moveWS(t, StatusRendered)
	o.ResumePDF, o.LetterPDF, o.TraceDir = "", "", ""
	res, err := Move(l, o)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Moved) != 2 || len(res.Skipped) != 0 {
		t.Errorf("moved %v skipped %v", res.Moved, res.Skipped)
	}
}

// A handoff failure after the move is reported but does not undo it.
func TestMoveHandoffFailureKeepsMove(t *testing.T) {
	l, o, _ := moveWS(t, StatusRendered)
	writeT(t, filepath.Join(l.Root, "not-a-dir"), "x", 0o644)
	l.Handoff = filepath.Join(l.Root, "not-a-dir")
	res, err := Move(l, o)
	if err == nil || res.Posting != "applied/"+mvSlug+".md" {
		t.Fatalf("got %+v, %v", res, err)
	}
	if _, serr := os.Stat(filepath.Join(l.Root, "applied", mvSlug+".md")); serr != nil {
		t.Errorf("move was undone: %v", serr)
	}
}

func TestClearHandoffMissing(t *testing.T) {
	if err := ClearHandoff(filepath.Join(t.TempDir(), "none")); err != nil {
		t.Error(err)
	}
}
