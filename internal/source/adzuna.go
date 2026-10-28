package source

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// adzunaBase is the Adzuna search API; the path continues with
// /<country>/search/1.
var adzunaBase = "https://api.adzuna.com/v1/api/jobs"

// adzunaPerPage is the page size of each Adzuna request.
const adzunaPerPage = 20

// adzunaCurrency maps an Adzuna country code to the currency its salaries use.
var adzunaCurrency = map[string]string{
	"at": "EUR", "be": "EUR", "de": "EUR", "es": "EUR", "fr": "EUR", "it": "EUR", "nl": "EUR",
	"gb": "GBP", "us": "USD", "ca": "CAD", "au": "AUD", "nz": "NZD", "ch": "CHF", "pl": "PLN",
	"br": "BRL", "in": "INR", "mx": "MXN", "sg": "SGD", "za": "ZAR",
}

type adzuna struct {
	client    Doer
	keys      Keys
	countries []string
}

func (a *adzuna) Name() string { return Adzuna }

type adzunaResponse struct {
	Results []struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Company struct {
			DisplayName string `json:"display_name"`
		} `json:"company"`
		Location struct {
			DisplayName string `json:"display_name"`
		} `json:"location"`
		URL     string  `json:"redirect_url"`
		Desc    string  `json:"description"`
		Created string  `json:"created"`
		Min     float64 `json:"salary_min"`
		Max     float64 `json:"salary_max"`
		Guess   string  `json:"salary_is_predicted"`
	} `json:"results"`
}

// Search sends one request per country and term (what_phrase).
func (a *adzuna) Search(ctx context.Context, terms []string) ([]Job, error) {
	if len(terms) == 0 {
		terms = []string{""}
	}
	var jobs []Job
	for _, country := range a.countries {
		for _, term := range terms {
			var resp adzunaResponse
			if err := getJSON(ctx, a.client, a.url(country, term), &resp); err != nil {
				return nil, err
			}
			jobs = append(jobs, a.jobs(country, resp)...)
		}
	}
	return jobs, nil
}

func (a *adzuna) url(country, term string) string {
	q := url.Values{}
	q.Set("app_id", a.keys.AdzunaAppID)
	q.Set("app_key", a.keys.AdzunaAppKey)
	q.Set("results_per_page", fmt.Sprint(adzunaPerPage))
	q.Set("content-type", "application/json")
	if term != "" {
		q.Set("what_phrase", term)
	}
	return fmt.Sprintf("%s/%s/search/1?%s", adzunaBase, url.PathEscape(strings.ToLower(country)), q.Encode())
}

func (a *adzuna) jobs(country string, resp adzunaResponse) []Job {
	out := make([]Job, 0, len(resp.Results))
	for _, r := range resp.Results {
		j := Job{
			ID: Adzuna + ":" + r.ID, Source: Adzuna, Title: r.Title,
			Company: r.Company.DisplayName, URL: r.URL, Location: r.Location.DisplayName,
			WorkMode: WorkMode(r.Title, r.Location.DisplayName, r.Desc),
			Deadline: Deadline(r.Desc), Posted: Date(r.Created), Summary: summarise(r.Desc),
		}
		// Adzuna estimates salaries it was not given; only stated ones count.
		if r.Guess != "1" && r.Min > 0 {
			j.SalaryMin, j.SalaryMax = int(r.Min), int(r.Max)
			j.SalaryCurrency, j.SalaryPeriod = adzunaCurrency[strings.ToLower(country)], "year"
		}
		out = append(out, j)
	}
	return out
}
