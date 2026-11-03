package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lamurian/caddie/internal/workspace"
)

// cmdFollowUps reports applied postings due a follow-up, open manual
// checklists and near or past deadlines (PRD §7, ADR 0027). The scan lives
// in workspace.FollowUps; the clock is e.today.
func cmdFollowUps(args []string, e env) error {
	fs := newFlags("follow-ups", e)
	asJSON := fs.Bool("json", false, "print the report as JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	today, err := time.Parse("2006-01-02", e.today())
	if err != nil {
		return fmt.Errorf("bad clock: %w", err)
	}
	_, layout, prefs, err := loadWorkspace(e)
	if err != nil {
		return err
	}
	r, err := workspace.FollowUps(layout, prefs.Apply.DeadlineWarningDays, today)
	if err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(e.stdout, r)
	}
	_, err = io.WriteString(e.stdout, followUpsText(r))
	return err
}

// followUpsText is the terse text form: one headed group per non-empty
// list, or a single line when there is nothing to do.
func followUpsText(r workspace.FollowUpReport) string {
	if len(r.FollowUps)+len(r.ManualSteps)+len(r.Deadlines) == 0 {
		return "Nothing to follow up.\n"
	}
	var b strings.Builder
	writeDue(&b, r)
	writeChecklists(&b, r.ManualSteps)
	writeDeadlines(&b, r)
	return b.String()
}

func writeDue(b *strings.Builder, r workspace.FollowUpReport) {
	if len(r.FollowUps) == 0 {
		return
	}
	fmt.Fprintf(b, "Follow-ups due (%s):\n", r.Today)
	for _, f := range r.FollowUps {
		fmt.Fprintf(b, "  %s  %s at %s  applied %s, follow up after %s (%s overdue)\n",
			f.Posting, f.Title, f.Company, f.AppliedAt, f.FollowUpAfter, days(f.DaysOverdue))
	}
}

func writeChecklists(b *strings.Builder, lists []workspace.OpenChecklist) {
	if len(lists) == 0 {
		return
	}
	b.WriteString("Open manual checklists:\n")
	for _, m := range lists {
		fmt.Fprintf(b, "  %s  %s at %s  %d open", m.Posting, m.Title, m.Company, m.Open)
		if m.ManualReason != "" {
			fmt.Fprintf(b, " (%s)", m.ManualReason)
		}
		b.WriteString("\n")
		for _, item := range m.Items {
			fmt.Fprintf(b, "    - [ ] %s\n", item)
		}
	}
}

func writeDeadlines(b *strings.Builder, r workspace.FollowUpReport) {
	if len(r.Deadlines) == 0 {
		return
	}
	fmt.Fprintf(b, "Deadlines (within %s):\n", days(r.DeadlineWarningDays))
	for _, d := range r.Deadlines {
		fmt.Fprintf(b, "  %s  %s at %s  deadline %s  %s\n",
			d.Posting, d.Title, d.Company, d.Deadline, deadlineNote(d))
	}
}

func deadlineNote(d workspace.DeadlineWarning) string {
	switch {
	case d.Past:
		return "PASSED " + days(-d.DaysLeft) + " ago"
	case d.DaysLeft == 0:
		return "today"
	default:
		return days(d.DaysLeft) + " left"
	}
}

func days(n int) string {
	if n == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", n)
}
