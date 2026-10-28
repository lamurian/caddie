package source

import (
	"context"
)

// arbeitnowBase is Arbeitnow's job board API (first page, newest first).
var arbeitnowBase = "https://www.arbeitnow.com/api/job-board-api"

type arbeitnow struct{ client Doer }

func (a *arbeitnow) Name() string { return Arbeitnow }

type arbeitnowResponse struct {
	Data []struct {
		Slug     string `json:"slug"`
		Company  string `json:"company_name"`
		Title    string `json:"title"`
		Desc     string `json:"description"`
		Remote   bool   `json:"remote"`
		URL      string `json:"url"`
		Location string `json:"location"`
		Created  int64  `json:"created_at"`
	} `json:"data"`
}

// Search fetches the newest jobs and keeps those matching a term, because
// Arbeitnow has no search parameter.
func (a *arbeitnow) Search(ctx context.Context, terms []string) ([]Job, error) {
	var resp arbeitnowResponse
	if err := getJSON(ctx, a.client, arbeitnowBase, &resp); err != nil {
		return nil, err
	}
	var jobs []Job
	for _, x := range resp.Data {
		text := Text(x.Desc)
		if !matches(terms, x.Title, text) {
			continue
		}
		mode := WorkMode(x.Title, x.Location, text)
		if x.Remote && mode != "hybrid" {
			mode = "wfh"
		}
		jobs = append(jobs, Job{
			ID: Arbeitnow + ":" + x.Slug, Source: Arbeitnow, Title: x.Title,
			Company: x.Company, URL: x.URL, Location: x.Location, WorkMode: mode,
			Deadline: Deadline(text), Posted: unixDate(x.Created), Summary: summarise(text),
		})
	}
	return jobs, nil
}
