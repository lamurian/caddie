package workspace

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/lamurian/caddie/internal/config"
)

// dateLayout is how every frontmatter date is written (PRD §12.3).
const dateLayout = "2006-01-02"

// FollowUpReport is what `caddie follow-ups` prints (ADR 0027). The three
// lists are never nil, so an empty report marshals to empty arrays.
type FollowUpReport struct {
	Today               string            `json:"today"`
	DeadlineWarningDays int               `json:"deadline_warning_days"`
	FollowUps           []FollowUpDue     `json:"follow_ups"`
	ManualSteps         []OpenChecklist   `json:"manual_steps"`
	Deadlines           []DeadlineWarning `json:"deadlines"`
}

// FollowUpDue is an applied/ posting whose follow_up_after has been reached.
// DaysOverdue is 0 on the follow_up_after day itself.
type FollowUpDue struct {
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Company       string `json:"company"`
	Posting       string `json:"posting"`
	AppliedAt     string `json:"applied_at"`
	FollowUpAfter string `json:"follow_up_after"`
	DaysOverdue   int    `json:"days_overdue"`
}

// OpenChecklist is a manual/ posting with unticked `- [ ]` items in its
// `## Manual steps` section. Open is len(Items).
type OpenChecklist struct {
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Company      string   `json:"company"`
	Posting      string   `json:"posting"`
	ManualReason string   `json:"manual_reason"`
	Open         int      `json:"open"`
	Items        []string `json:"items"`
}

// DeadlineWarning is a postings/ or manual/ posting whose deadline is at
// most DeadlineWarningDays away (DaysLeft 0 = today) or already gone
// (Past, DaysLeft < 0).
type DeadlineWarning struct {
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Company  string `json:"company"`
	Posting  string `json:"posting"`
	Status   Status `json:"status"`
	Deadline string `json:"deadline"`
	DaysLeft int    `json:"days_left"`
	Past     bool   `json:"past"`
}

// FollowUps scans applied/, manual/ and postings/ for what needs attention
// on the calendar day of today (its own location; the clock is injected).
// Missing directories count as empty. A malformed posting or an
// unparseable date wraps ErrBadPosting and names the file; a negative
// warnDays wraps config.ErrConfig.
func FollowUps(l Layout, warnDays int, today time.Time) (FollowUpReport, error) {
	if warnDays < 0 {
		return FollowUpReport{}, fmt.Errorf("%w: apply.deadline_warning_days %d is negative",
			config.ErrConfig, warnDays)
	}
	day := civil(today)
	r := FollowUpReport{
		Today: day.Format(dateLayout), DeadlineWarningDays: warnDays,
		FollowUps: []FollowUpDue{}, ManualSteps: []OpenChecklist{}, Deadlines: []DeadlineWarning{},
	}
	for _, dir := range []string{DirApplied, DirManual, DirPostings} {
		files, err := readPostings(l, dir)
		if err != nil {
			return FollowUpReport{}, err
		}
		for _, f := range files {
			if err := r.add(dir, f, day, warnDays); err != nil {
				return FollowUpReport{}, err
			}
		}
	}
	r.sort()
	return r, nil
}

// add files one posting under every category it belongs to: follow-ups
// for applied/, open steps and deadlines for manual/, deadlines for
// postings/.
func (r *FollowUpReport) add(dir string, f postingFile, day time.Time, warnDays int) error {
	if dir == DirApplied {
		return r.addFollowUp(f, day)
	}
	if dir == DirManual {
		r.addSteps(f)
	}
	return r.addDeadline(f, day, warnDays)
}

func (r *FollowUpReport) addFollowUp(f postingFile, day time.Time) error {
	p := f.posting
	due, ok, err := dateField(f, "follow_up_after", p.FollowUpAfter)
	if err != nil || !ok || day.Before(due) {
		return err
	}
	r.FollowUps = append(r.FollowUps, FollowUpDue{Slug: p.Slug, Title: p.Title,
		Company: p.Company, Posting: f.rel, AppliedAt: p.AppliedAt,
		FollowUpAfter: p.FollowUpAfter, DaysOverdue: daysBetween(due, day)})
	return nil
}

func (r *FollowUpReport) addSteps(f postingFile) {
	p := f.posting
	if open := openSteps(f.body); len(open) > 0 {
		r.ManualSteps = append(r.ManualSteps, OpenChecklist{Slug: p.Slug, Title: p.Title,
			Company: p.Company, Posting: f.rel, ManualReason: p.ManualReason,
			Open: len(open), Items: open})
	}
}

func (r *FollowUpReport) addDeadline(f postingFile, day time.Time, warnDays int) error {
	p := f.posting
	end, ok, err := dateField(f, "deadline", p.Deadline)
	if err != nil || !ok {
		return err
	}
	if left := daysBetween(day, end); left <= warnDays {
		r.Deadlines = append(r.Deadlines, DeadlineWarning{Slug: p.Slug, Title: p.Title,
			Company: p.Company, Posting: f.rel, Status: p.Status, Deadline: p.Deadline,
			DaysLeft: left, Past: left < 0})
	}
	return nil
}

// sort orders each list by its date (most urgent first), then by path.
func (r *FollowUpReport) sort() {
	slices.SortFunc(r.FollowUps, func(a, b FollowUpDue) int {
		return cmp.Or(cmp.Compare(a.FollowUpAfter, b.FollowUpAfter), cmp.Compare(a.Posting, b.Posting))
	})
	slices.SortFunc(r.ManualSteps, func(a, b OpenChecklist) int {
		return cmp.Compare(a.Posting, b.Posting)
	})
	slices.SortFunc(r.Deadlines, func(a, b DeadlineWarning) int {
		return cmp.Or(cmp.Compare(a.Deadline, b.Deadline), cmp.Compare(a.Posting, b.Posting))
	})
}

// dateField parses a YYYY-MM-DD frontmatter value; ok is false when unset.
func dateField(f postingFile, field, value string) (time.Time, bool, error) {
	if value == "" {
		return time.Time{}, false, nil
	}
	t, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("%w: %s: %s %q is not a YYYY-MM-DD date",
			ErrBadPosting, f.rel, field, value)
	}
	return t, true, nil
}

// civil is the calendar day of t as midnight UTC, so day counts ignore
// time zones and daylight saving.
func civil(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// daysBetween counts whole days from a to b (both civil days).
func daysBetween(a, b time.Time) int { return int(b.Sub(a).Hours() / 24) }
