package workspace

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

var isoDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// Frontmatter renders p as the "---"-fenced YAML block of PRD §12.3: every
// field in PRD order, unset values left empty (`applied_at:`), dates
// unquoted and other strings quoted only where YAML needs it.
func (p Posting) Frontmatter() []byte {
	fields := []struct{ key, val string }{
		{"slug", scalar(p.Slug)}, {"title", scalar(p.Title)}, {"company", scalar(p.Company)},
		{"url", scalar(p.URL)}, {"source", scalar(p.Source)}, {"ats", scalar(p.ATS)},
		{"ats_host", scalar(p.ATSHost)}, {"location", scalar(p.Location)},
		{"work_mode", scalar(p.WorkMode)}, {"salary_min", number(p.SalaryMin)},
		{"salary_max", number(p.SalaryMax)}, {"salary_currency", scalar(p.SalaryCurrency)},
		{"salary_period", scalar(p.SalaryPeriod)}, {"deadline", date(p.Deadline)},
		{"status", scalar(string(p.Status))}, {"saved_at", date(p.SavedAt)},
		{"applied_at", date(p.AppliedAt)}, {"follow_up_after", date(p.FollowUpAfter)},
		{"manual_reason", scalar(p.ManualReason)},
	}
	var b bytes.Buffer
	b.WriteString("---\n")
	for _, f := range fields {
		if f.val == "" {
			fmt.Fprintf(&b, "%s:\n", f.key)
			continue
		}
		fmt.Fprintf(&b, "%s: %s\n", f.key, f.val)
	}
	b.WriteString("---\n")
	return b.Bytes()
}

// scalar renders one line of text as a YAML scalar, collapsing whitespace.
func scalar(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return ""
	}
	out, err := yaml.Marshal(s)
	if err != nil {
		return strconv.Quote(s)
	}
	return strings.TrimSuffix(string(out), "\n")
}

func number(n int) string {
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}

func date(s string) string {
	if isoDate.MatchString(s) {
		return s
	}
	return scalar(s)
}
