package workspace

import (
	"bytes"
	"errors"
	"fmt"

	"go.yaml.in/yaml/v3"
)

// ErrBadPosting marks a posting file whose frontmatter is invalid.
var ErrBadPosting = errors.New("invalid posting")

// Posting is the YAML frontmatter of postings/<slug>.md (PRD §12.3).
// Dates are kept as written (YYYY-MM-DD); empty means unset.
type Posting struct {
	Slug           string `yaml:"slug" json:"slug"`
	Title          string `yaml:"title" json:"title"`
	Company        string `yaml:"company" json:"company"`
	URL            string `yaml:"url" json:"url"`
	Source         string `yaml:"source" json:"source"`
	ATS            string `yaml:"ats" json:"ats"`
	ATSHost        string `yaml:"ats_host" json:"ats_host"`
	Location       string `yaml:"location" json:"location"`
	WorkMode       string `yaml:"work_mode" json:"work_mode"`
	SalaryMin      int    `yaml:"salary_min" json:"salary_min"`
	SalaryMax      int    `yaml:"salary_max" json:"salary_max"`
	SalaryCurrency string `yaml:"salary_currency" json:"salary_currency"`
	SalaryPeriod   string `yaml:"salary_period" json:"salary_period"`
	Deadline       string `yaml:"deadline" json:"deadline"`
	Status         Status `yaml:"status" json:"status"`
	SavedAt        string `yaml:"saved_at" json:"saved_at"`
	AppliedAt      string `yaml:"applied_at" json:"applied_at"`
	FollowUpAfter  string `yaml:"follow_up_after" json:"follow_up_after"`
	ManualReason   string `yaml:"manual_reason" json:"manual_reason"`
}

var fence = []byte("---")

// ParsePosting splits a posting file into its frontmatter and Markdown body
// and validates the status. name is used in error messages.
func ParsePosting(name string, data []byte) (Posting, []byte, error) {
	front, body, ok := splitFrontmatter(data)
	if !ok {
		return Posting{}, nil, fmt.Errorf("%w: %s:1: missing --- frontmatter", ErrBadPosting, name)
	}
	var p Posting
	dec := yaml.NewDecoder(bytes.NewReader(append([]byte("\n"), front...)))
	dec.KnownFields(true)
	if err := dec.Decode(&p); err != nil {
		return Posting{}, nil, fmt.Errorf("%w: %s: %v", ErrBadPosting, name, err)
	}
	if p.Slug == "" {
		return Posting{}, nil, fmt.Errorf("%w: %s: slug is empty", ErrBadPosting, name)
	}
	if _, err := ParseStatus(string(p.Status)); err != nil {
		return Posting{}, nil, fmt.Errorf("%w: %s: %w", ErrBadPosting, name, err)
	}
	return p, body, nil
}

// splitFrontmatter returns the YAML between a leading "---" line and the
// next "---" line, and the rest of the file.
func splitFrontmatter(data []byte) (front, body []byte, ok bool) {
	lines := bytes.SplitAfter(data, []byte("\n"))
	if len(lines) == 0 || !bytes.Equal(bytes.TrimSpace(lines[0]), fence) {
		return nil, nil, false
	}
	for i := 1; i < len(lines); i++ {
		if bytes.Equal(bytes.TrimSpace(lines[i]), fence) {
			return bytes.Join(lines[1:i], nil), bytes.Join(lines[i+1:], nil), true
		}
	}
	return nil, nil, false
}
