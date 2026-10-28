package source

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"strings"
)

// leverBase is the Lever postings API; the path continues with /<board>.
var leverBase = "https://api.lever.co/v0/postings"

type lever struct {
	client Doer
	boards []string
}

func (l *lever) Name() string { return Lever }

type leverJob struct {
	ID         string `json:"id"`
	Title      string `json:"text"`
	URL        string `json:"hostedUrl"`
	Created    int64  `json:"createdAt"`
	Workplace  string `json:"workplaceType"`
	Desc       string `json:"description"`
	Additional string `json:"additional"`
	Categories struct {
		Location string `json:"location"`
	} `json:"categories"`
	Lists []struct {
		Text    string `json:"text"`
		Content string `json:"content"`
	} `json:"lists"`
	Salary *struct {
		Min      float64 `json:"min"`
		Max      float64 `json:"max"`
		Currency string  `json:"currency"`
		Interval string  `json:"interval"`
	} `json:"salaryRange"`
}

// Search fetches every posting on each board and keeps those matching a term.
func (l *lever) Search(ctx context.Context, terms []string) ([]Job, error) {
	var jobs []Job
	for _, board := range l.boards {
		var resp []leverJob
		u := fmt.Sprintf("%s/%s?mode=json", leverBase, url.PathEscape(board))
		if err := getJSON(ctx, l.client, u, &resp); err != nil {
			return nil, err
		}
		for _, x := range resp {
			j, text := x.job(board)
			if matches(terms, x.Title, text) {
				jobs = append(jobs, j)
			}
		}
	}
	return jobs, nil
}

var leverModes = map[string]string{"remote": "wfh", "hybrid": "hybrid", "on-site": "wfo", "onsite": "wfo"}

var leverPeriods = map[string]string{"per-year-salary": "year", "per-month-salary": "month",
	"per-hour-wage": "hour"}

// job normalises a Lever posting and returns it with its plain text.
// Lever does not name the company, so the board name stands in for it.
func (x leverJob) job(board string) (Job, string) {
	text := Text(x.HTML())
	mode := leverModes[x.Workplace]
	if mode == "" {
		mode = WorkMode(x.Categories.Location, x.Title)
	}
	j := Job{
		ID: Lever + ":" + x.ID, Source: Lever, Title: x.Title, Company: board, URL: x.URL,
		Location: x.Categories.Location, WorkMode: mode, Deadline: Deadline(text),
		Posted: unixDate(x.Created / 1000), Summary: summarise(text),
	}
	if s := x.Salary; s != nil && s.Min > 0 {
		Salary{Min: int(s.Min), Max: int(s.Max), Currency: s.Currency, Period: leverPeriods[s.Interval]}.apply(&j)
	}
	return j, text
}

// HTML joins the description, the titled lists and the closing text.
func (x leverJob) HTML() string {
	var b strings.Builder
	b.WriteString(x.Desc)
	for _, l := range x.Lists {
		fmt.Fprintf(&b, "<h3>%s</h3><ul>%s</ul>", html.EscapeString(l.Text), l.Content)
	}
	b.WriteString(x.Additional)
	return b.String()
}
