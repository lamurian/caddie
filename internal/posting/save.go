package posting

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/lamurian/caddie/internal/memory"
	"github.com/lamurian/caddie/internal/source"
	"github.com/lamurian/caddie/internal/workspace"
)

// ErrNoPosting means the URL did not lead to a recognisable job posting
// (exit code 5 in the CLI).
var ErrNoPosting = errors.New("no job posting found")

// Options carries everything Save touches, so tests can fake it.
type Options struct {
	Layout workspace.Layout
	Store  memory.Store
	Client source.Doer
	Today  string // YYYY-MM-DD, written as saved_at and in seen.jsonl
}

// Result is what `caddie posting save` prints.
type Result struct {
	Slug    string   `json:"slug"`
	Path    string   `json:"path"`
	ID      string   `json:"id"`
	Created bool     `json:"created"`
	Review  bool     `json:"review"`
	Missing []string `json:"missing_sections"`
}

// Save fetches the posting at rawURL and writes postings/<slug>.md with
// status saved, copying reviews/<company>.md into ## Reviews, then marks
// the job saved in seen.jsonl. Saving a URL that is already saved (in
// postings/, applied/ or manual/) rewrites nothing and returns that slug.
func Save(ctx context.Context, rawURL string, o Options) (Result, error) {
	if res, p, ok := findByURL(o.Layout, rawURL); ok {
		j, err := o.merge(source.Job{ID: source.ID(rawURL), Title: p.Title, Company: p.Company}, rawURL)
		if err != nil {
			return Result{}, err
		}
		res.ID = j.ID
		return res, o.Store.Record(memory.Entry{ID: j.ID, URL: rawURL, Title: p.Title,
			Company: p.Company, Status: memory.Saved, At: o.Today})
	}
	d, err := source.Fetch(ctx, o.Client, rawURL)
	if err != nil {
		return Result{}, err
	}
	fetchedTitle := d.Job.Title
	job, err := o.merge(d.Job, rawURL)
	if err != nil {
		return Result{}, err
	}
	if job.Title == "" {
		return Result{}, fmt.Errorf("%w at %s: the page states no job title", ErrNoPosting, rawURL)
	}
	slug := freeSlug(o.Layout, workspace.PostingSlug(job.Title, job.Company))
	review, reviewFile, err := findReview(o.Layout, job.Company)
	if err != nil {
		return Result{}, err
	}
	secs := Extract(d.HTML, job.Title, fetchedTitle)
	p := frontmatter(job, slug, rawURL, o.Today)
	content := string(p.Frontmatter()) + "\n" + secs.Markdown() + reviewsSection(review, job.Company)
	path := filepath.Join(o.Layout.Dir(workspace.DirPostings), slug+".md")
	if err := writeNew(path, strings.TrimRight(content, "\n")+"\n"); err != nil {
		return Result{}, err
	}
	res := Result{Slug: slug, Path: path, ID: job.ID, Created: true, Review: reviewFile != "",
		Missing: secs.Missing()}
	return res, o.Store.Record(memory.Entry{ID: job.ID, URL: rawURL, Title: job.Title,
		Company: job.Company, Status: memory.Saved, At: o.Today})
}

// merge completes the fetched job with what the last search or seen.jsonl
// knows about the URL. A remembered ID, source, title and company win, so
// the posting matches what the agent showed the user.
func (o Options) merge(j source.Job, rawURL string) (source.Job, error) {
	known, ok, err := o.Store.RecallURL(rawURL)
	if err != nil {
		return j, err
	}
	if !ok {
		log, err := o.Store.Log()
		if err != nil {
			return j, err
		}
		e, found := log.ByURL(rawURL)
		if !found {
			return j, nil
		}
		known = source.Job{ID: e.ID, Title: e.Title, Company: e.Company}
		if prefix, _, _ := strings.Cut(e.ID, ":"); source.Known(prefix) {
			known.Source = prefix
		}
	}
	j.ID, j.Source = firstSet(known.ID, j.ID), firstSet(known.Source, j.Source)
	j.Title, j.Company = firstSet(known.Title, j.Title), firstSet(known.Company, j.Company)
	j.Location, j.WorkMode = firstSet(j.Location, known.Location), firstSet(j.WorkMode, known.WorkMode)
	j.Deadline = firstSet(j.Deadline, known.Deadline)
	if j.SalaryMin == 0 && known.SalaryMin > 0 {
		j.SalaryMin, j.SalaryMax, j.SalaryCurrency, j.SalaryPeriod =
			known.SalaryMin, known.SalaryMax, known.SalaryCurrency, known.SalaryPeriod
	}
	return j, nil
}

func firstSet(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// atsHosts maps an applicant tracking system's host suffix to its name.
var atsHosts = [][2]string{{"greenhouse.io", "greenhouse"}, {"lever.co", "lever"},
	{"myworkdayjobs.com", "workday"}, {"smartrecruiters.com", "smartrecruiters"},
	{"ashbyhq.com", "ashby"}, {"recruitee.com", "recruitee"}, {"personio.de", "personio"},
	{"personio.com", "personio"}, {"teamtailor.com", "teamtailor"}, {"workable.com", "workable"}}

func frontmatter(j source.Job, slug, rawURL, today string) workspace.Posting {
	p := workspace.Posting{
		Slug: slug, Title: j.Title, Company: j.Company, URL: rawURL, Source: j.Source,
		Location: j.Location, WorkMode: j.WorkMode, SalaryMin: j.SalaryMin, SalaryMax: j.SalaryMax,
		SalaryCurrency: j.SalaryCurrency, SalaryPeriod: j.SalaryPeriod, Deadline: j.Deadline,
		Status: workspace.StatusSaved, SavedAt: today,
	}
	if u, err := url.Parse(rawURL); err == nil {
		p.ATSHost = strings.ToLower(u.Hostname())
		for _, a := range atsHosts {
			if p.ATSHost == a[0] || strings.HasSuffix(p.ATSHost, "."+a[0]) {
				p.ATS = a[1]
			}
		}
	}
	return p
}

// writeNew creates path, failing rather than overwriting an existing file.
func writeNew(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(content); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
