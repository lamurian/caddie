package source

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// remotiveBase is Remotive's public remote-jobs API.
var remotiveBase = "https://remotive.com/api/remote-jobs"

type remotive struct{ client Doer }

func (r *remotive) Name() string { return Remotive }

type remotiveResponse struct {
	Jobs []struct {
		ID       int    `json:"id"`
		URL      string `json:"url"`
		Title    string `json:"title"`
		Company  string `json:"company_name"`
		Location string `json:"candidate_required_location"`
		Salary   string `json:"salary"`
		Date     string `json:"publication_date"`
		Desc     string `json:"description"`
	} `json:"jobs"`
}

// Search sends one request per term. Remotive's search matches loosely, so
// a job is kept only when a term occurs in its title or description.
func (r *remotive) Search(ctx context.Context, terms []string) ([]Job, error) {
	queries := terms
	if len(queries) == 0 {
		queries = []string{""}
	}
	var jobs []Job
	for _, term := range queries {
		u := remotiveBase
		if term != "" {
			u += "?" + url.Values{"search": {term}}.Encode()
		}
		var resp remotiveResponse
		if err := getJSON(ctx, r.client, u, &resp); err != nil {
			return nil, err
		}
		for _, x := range resp.Jobs {
			text := Text(x.Desc)
			if !matches(terms, x.Title, text) {
				continue
			}
			j := Job{
				ID: fmt.Sprintf("%s:%d", Remotive, x.ID), Source: Remotive, Title: x.Title,
				Company: x.Company, URL: x.URL, Location: x.Location,
				WorkMode: remoteMode(x.Location), Deadline: Deadline(text),
				Posted: Date(x.Date), Summary: summarise(text),
			}
			ParseSalary(x.Salary).apply(&j)
			jobs = append(jobs, j)
		}
	}
	return jobs, nil
}

// remoteMode is the work mode of a remote-only job: wfa when it may be done
// from anywhere, else wfh (remote within a region).
func remoteMode(location string) string {
	l := strings.ToLower(location)
	if l == "" || strings.Contains(l, "worldwide") || strings.Contains(l, "anywhere") {
		return "wfa"
	}
	return "wfh"
}
