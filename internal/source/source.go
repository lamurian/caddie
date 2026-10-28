package source

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrFetch marks a network or upstream failure (exit code 8 in the CLI).
var ErrFetch = errors.New("fetch failed")

// Source names, as used by `caddie search --source` and in Job IDs.
const (
	Adzuna     = "adzuna"
	Remotive   = "remotive"
	Arbeitnow  = "arbeitnow"
	Greenhouse = "greenhouse-board"
	Lever      = "lever-board"
)

// Names lists every board source in the order search queries them.
var Names = []string{Adzuna, Remotive, Arbeitnow, Greenhouse, Lever}

// Job is one posting normalised from any source. ID is "<source>:<native
// id>" (ADR 0016). Optional fields are omitted when the source has no value.
type Job struct {
	ID             string `json:"id"`
	Source         string `json:"source"`
	Title          string `json:"title"`
	Company        string `json:"company"`
	URL            string `json:"url"`
	Location       string `json:"location,omitempty"`
	WorkMode       string `json:"work_mode,omitempty"`
	SalaryMin      int    `json:"salary_min,omitempty"`
	SalaryMax      int    `json:"salary_max,omitempty"`
	SalaryCurrency string `json:"salary_currency,omitempty"`
	SalaryPeriod   string `json:"salary_period,omitempty"`
	Deadline       string `json:"deadline,omitempty"`
	Posted         string `json:"posted,omitempty"`
	Summary        string `json:"summary,omitempty"`
}

// Source is one job board. Search returns the jobs matching any of terms
// (all jobs when terms is empty), normalised to Job. A source with several
// boards may return jobs together with a *partial error listing the boards
// it skipped.
type Source interface {
	Name() string
	Search(ctx context.Context, terms []string) ([]Job, error)
}

// Known reports whether name is a board source.
func Known(name string) bool {
	for _, n := range Names {
		if n == name {
			return true
		}
	}
	return false
}

// Filter is applied to the combined results of every source.
type Filter struct {
	ExcludeRoles     []string        // title substrings, case-insensitive
	ExcludeCompanies []string        // exact company names, case-insensitive
	Seen             map[string]bool // IDs to hide; nil shows everything
	Limit            int             // 0 means no limit
}

// Result is what `caddie search` prints.
type Result struct {
	Jobs       []Job    `json:"jobs"`
	HiddenSeen int      `json:"hidden_seen"`
	Notices    []string `json:"notices"`
}

// Search queries every source in order, de-duplicates by URL (first wins)
// and applies f. A failing source becomes a notice, as does each skipped
// board of a partial result; only when every source fails is the result an
// ErrFetch error.
func Search(ctx context.Context, srcs []Source, terms []string, f Filter) (Result, error) {
	res := Result{Jobs: []Job{}, Notices: []string{}}
	urls := map[string]bool{}
	failed := 0
	for _, s := range srcs {
		jobs, err := s.Search(ctx, terms)
		var p *partial
		if errors.As(err, &p) {
			res.Notices = append(res.Notices, p.notices...)
			err = nil
		}
		if err != nil {
			failed++
			res.Notices = append(res.Notices, fmt.Sprintf("%s: skipped: %v", s.Name(), err))
			continue
		}
		res.add(jobs, f, urls)
	}
	if len(srcs) > 0 && failed == len(srcs) {
		return res, fmt.Errorf("%w: every source failed: %s", ErrFetch, strings.Join(res.Notices, "; "))
	}
	if f.Limit > 0 && len(res.Jobs) > f.Limit {
		res.Jobs = res.Jobs[:f.Limit]
	}
	return res, nil
}

// add appends the jobs f keeps whose URL is new, counting hidden seen ones.
func (res *Result) add(jobs []Job, f Filter, urls map[string]bool) {
	for _, j := range jobs {
		if urls[j.URL] || !f.keeps(j) {
			continue
		}
		urls[j.URL] = true
		if f.Seen[j.ID] {
			res.HiddenSeen++
			continue
		}
		res.Jobs = append(res.Jobs, j)
	}
}

func (f Filter) keeps(j Job) bool {
	title := strings.ToLower(j.Title)
	for _, r := range f.ExcludeRoles {
		if r != "" && strings.Contains(title, strings.ToLower(r)) {
			return false
		}
	}
	for _, c := range f.ExcludeCompanies {
		if strings.EqualFold(strings.TrimSpace(c), strings.TrimSpace(j.Company)) {
			return false
		}
	}
	return true
}

// matches reports whether any term occurs in the texts (case-insensitive).
// Sources without a server-side search use it; no terms matches everything.
func matches(terms []string, texts ...string) bool {
	if len(terms) == 0 {
		return true
	}
	hay := strings.ToLower(strings.Join(texts, " "))
	for _, t := range terms {
		if t != "" && strings.Contains(hay, strings.ToLower(t)) {
			return true
		}
	}
	return false
}
