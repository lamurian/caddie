package source

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// greenhouseBase is the Greenhouse job board API; the path continues with
// /<board>/jobs.
var greenhouseBase = "https://boards-api.greenhouse.io/v1/boards"

type greenhouse struct {
	client Doer
	boards []string
}

func (g *greenhouse) Name() string { return Greenhouse }

type greenhouseJob struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	URL      string `json:"absolute_url"`
	Company  string `json:"company_name"`
	Updated  string `json:"updated_at"`
	First    string `json:"first_published"`
	Content  string `json:"content"`
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Departments []struct {
		Name string `json:"name"`
	} `json:"departments"`
	Metadata []struct {
		Name  string `json:"name"`
		Value any    `json:"value"`
	} `json:"metadata"`
	Pay []struct {
		Min      int64  `json:"min_cents"`
		Max      int64  `json:"max_cents"`
		Currency string `json:"currency_type"`
	} `json:"pay_input_ranges"`
}

// Search fetches every job on each board and keeps those whose title or
// department matches a term. The description is not searched: every job on
// a company board repeats the same "About us" text (ADR 0018). A failing
// board is skipped with a notice (see searchBoards).
func (g *greenhouse) Search(ctx context.Context, terms []string) ([]Job, error) {
	return searchBoards(Greenhouse, "sources.greenhouse_boards", g.boards, func(board string) ([]Job, error) {
		var resp struct {
			Jobs []greenhouseJob `json:"jobs"`
		}
		u := fmt.Sprintf("%s/%s/jobs?content=true", greenhouseBase, url.PathEscape(board))
		if err := getJSON(ctx, g.client, u, &resp); err != nil {
			return nil, err
		}
		var jobs []Job
		for _, x := range resp.Jobs {
			if matches(terms, x.labels()...) {
				jobs = append(jobs, x.job(board))
			}
		}
		return jobs, nil
	})
}

// job normalises a Greenhouse job.
func (x greenhouseJob) job(board string) Job {
	text := Text(x.Content)
	j := Job{
		ID: fmt.Sprintf("%s:%d", Greenhouse, x.ID), Source: Greenhouse, Title: x.Title,
		Company: x.Company, URL: x.URL, Location: x.Location.Name,
		WorkMode: WorkMode(x.Location.Name, x.Title), Deadline: x.deadline(text),
		Posted: Date(firstNonEmpty(x.First, x.Updated)), Summary: summarise(text),
	}
	if j.Company == "" {
		j.Company = board
	}
	if len(x.Pay) > 0 {
		p := x.Pay[0]
		Salary{Min: int(p.Min / 100), Max: int(p.Max / 100), Currency: p.Currency, Period: "year"}.apply(&j)
	}
	return j
}

// labels are the fields search terms are matched against: the title and
// the department names.
func (x greenhouseJob) labels() []string {
	out := []string{x.Title}
	for _, d := range x.Departments {
		out = append(out, d.Name)
	}
	return out
}

// deadline prefers a custom field named like "Application deadline".
func (x greenhouseJob) deadline(text string) string {
	for _, m := range x.Metadata {
		if v, ok := m.Value.(string); ok && strings.Contains(strings.ToLower(m.Name), "deadline") {
			if d := Date(v); d != "" {
				return d
			}
		}
	}
	return Deadline(text)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
