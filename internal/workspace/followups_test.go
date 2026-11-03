package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lamurian/caddie/internal/config"
)

// fuToday is the fixed clock: late evening in a zone east of UTC, so the
// calendar day must come from its own location, not from UTC.
var fuToday = time.Date(2026, 11, 3, 23, 30, 0, 0, time.FixedZone("UTC+9", 9*3600))

// fuWS returns a layout over an empty temp workspace (no directories).
func fuWS(t *testing.T) Layout {
	t.Helper()
	return Layout{Root: t.TempDir()}
}

// fuPost writes <dir>/<slug>.md with the given frontmatter fields and body.
func fuPost(t *testing.T, l Layout, dir string, p Posting, body string) {
	t.Helper()
	p.Title, p.Company = "Role "+p.Slug, "Org"
	if p.Status == "" {
		p.Status = StatusSaved
	}
	path := filepath.Join(l.Dir(dir), p.Slug+".md")
	mkdirT(t, filepath.Dir(path))
	writeT(t, path, string(p.Frontmatter())+body, 0o644)
}

func fuRun(t *testing.T, l Layout, warn int) FollowUpReport {
	t.Helper()
	r, err := FollowUps(l, warn, fuToday)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestFollowUpsEmpty(t *testing.T) {
	l := fuWS(t)
	r := fuRun(t, l, 7)
	out, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"today":"2026-11-03","deadline_warning_days":7,"follow_ups":[],"manual_steps":[],"deadlines":[]}`
	if string(out) != want {
		t.Errorf("empty report = %s, want %s", out, want)
	}
	for _, d := range []string{DirApplied, DirManual, DirPostings} {
		mkdirT(t, filepath.Join(l.Dir(d), "some-slug"))
		writeT(t, filepath.Join(l.Dir(d), "README.md"), "# not a posting", 0o644)
		writeT(t, filepath.Join(l.Dir(d), ".gitignore"), "*", 0o644)
	}
	if again := fuRun(t, l, 7); !reflect.DeepEqual(again, r) {
		t.Errorf("README/.gitignore/artefact dirs not skipped: %+v", again)
	}
}

func TestFollowUpsApplied(t *testing.T) {
	l := fuWS(t)
	st := StatusApplied
	fuPost(t, l, DirApplied, Posting{Slug: "b-due-today", Status: st, AppliedAt: "2026-10-13", FollowUpAfter: "2026-11-03"}, "")
	fuPost(t, l, DirApplied, Posting{Slug: "a-overdue", Status: st, AppliedAt: "2026-10-01", FollowUpAfter: "2026-10-22"}, "")
	fuPost(t, l, DirApplied, Posting{Slug: "c-tomorrow", Status: st, FollowUpAfter: "2026-11-04"}, "")
	fuPost(t, l, DirApplied, Posting{Slug: "d-unset", Status: st, Deadline: "2026-11-01"}, "")
	r := fuRun(t, l, 7)
	want := []FollowUpDue{
		{Slug: "a-overdue", Title: "Role a-overdue", Company: "Org", Posting: "applied/a-overdue.md",
			AppliedAt: "2026-10-01", FollowUpAfter: "2026-10-22", DaysOverdue: 12},
		{Slug: "b-due-today", Title: "Role b-due-today", Company: "Org", Posting: "applied/b-due-today.md",
			AppliedAt: "2026-10-13", FollowUpAfter: "2026-11-03", DaysOverdue: 0},
	}
	if !reflect.DeepEqual(r.FollowUps, want) {
		t.Errorf("follow_ups = %+v\nwant %+v", r.FollowUps, want)
	}
	if len(r.Deadlines) != 0 {
		t.Errorf("applied/ deadlines must not be reported: %+v", r.Deadlines)
	}
}

func TestFollowUpsManualSteps(t *testing.T) {
	l := fuWS(t)
	body := "\n## Description\n\n- [ ] not a step\n\n## Manual steps\n\n" +
		"- [x] Create an account\n- [ ] Upload the resume  \r\n* [ ] Answer the essay\n" +
		"\n## Notes\n\n- [ ] after the section\n"
	fuPost(t, l, DirManual, Posting{Slug: "open", Status: StatusManual, ManualReason: "CAPTCHA"}, body)
	fuPost(t, l, DirManual, Posting{Slug: "all-done", Status: StatusManual},
		"\n## Manual steps\n\n- [x] Done\n- [X] Also done\n")
	fuPost(t, l, DirManual, Posting{Slug: "no-section", Status: StatusManual}, "\n- [ ] loose item\n")
	r := fuRun(t, l, 7)
	want := []OpenChecklist{{Slug: "open", Title: "Role open", Company: "Org",
		Posting: "manual/open.md", ManualReason: "CAPTCHA", Open: 2,
		Items: []string{"Upload the resume", "Answer the essay"}}}
	if !reflect.DeepEqual(r.ManualSteps, want) {
		t.Errorf("manual_steps = %+v\nwant %+v", r.ManualSteps, want)
	}
}

func TestFollowUpsDeadlines(t *testing.T) {
	l := fuWS(t)
	fuPost(t, l, DirPostings, Posting{Slug: "edge", Deadline: "2026-11-10"}, "")   // today + 7
	fuPost(t, l, DirPostings, Posting{Slug: "beyond", Deadline: "2026-11-11"}, "") // today + 8
	fuPost(t, l, DirPostings, Posting{Slug: "today", Deadline: "2026-11-03"}, "")  // 0
	fuPost(t, l, DirPostings, Posting{Slug: "gone", Deadline: "2026-11-02"}, "")   // -1
	fuPost(t, l, DirManual, Posting{Slug: "man", Status: StatusManual, Deadline: "2026-11-05"}, "")
	fuPost(t, l, DirPostings, Posting{Slug: "none"}, "")
	got := map[string][2]any{}
	var order []string
	for _, d := range fuRun(t, l, 7).Deadlines {
		got[d.Slug] = [2]any{d.DaysLeft, d.Past}
		order = append(order, d.Slug)
	}
	want := map[string][2]any{"gone": {-1, true}, "today": {0, false}, "man": {2, false}, "edge": {7, false}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("deadlines = %v, want %v", got, want)
	}
	if strings.Join(order, ",") != "gone,today,man,edge" {
		t.Errorf("order = %v", order)
	}
	var zero []string
	for _, d := range fuRun(t, l, 0).Deadlines {
		zero = append(zero, d.Slug)
	}
	if strings.Join(zero, ",") != "gone,today" {
		t.Errorf("warn 0 = %v, want gone,today", zero)
	}
}

func TestFollowUpsErrors(t *testing.T) {
	l := fuWS(t)
	if _, err := FollowUps(l, -1, fuToday); !errors.Is(err, config.ErrConfig) {
		t.Errorf("negative warn days: %v", err)
	}
	fuPost(t, l, DirPostings, Posting{Slug: "bad-date", Deadline: "14 Nov 2026"}, "")
	_, err := FollowUps(l, 7, fuToday)
	if !errors.Is(err, ErrBadPosting) || !strings.Contains(err.Error(), "postings/bad-date.md") ||
		!strings.Contains(err.Error(), "deadline") {
		t.Errorf("bad date: %v", err)
	}
	if err := os.Remove(filepath.Join(l.Dir(DirPostings), "bad-date.md")); err != nil {
		t.Fatal(err)
	}
	mkdirT(t, l.Dir(DirApplied))
	writeT(t, filepath.Join(l.Dir(DirApplied), "broken.md"), "no frontmatter\n", 0o644)
	_, err = FollowUps(l, 7, fuToday)
	if !errors.Is(err, ErrBadPosting) || !strings.Contains(err.Error(), "applied/broken.md") {
		t.Errorf("malformed posting: %v", err)
	}
}
