package source

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

var fetchCases = []struct {
	name, url, api, file string
	want                 Job
	html                 string // a fragment the description HTML must contain
}{
	{"greenhouse", "https://boards.greenhouse.io/acme/jobs/123",
		"https://boards-api.greenhouse.io/v1/boards/acme/jobs/123?pay_transparency=true", "greenhouse-acme-123.json",
		Job{ID: "greenhouse-board:123", Source: Greenhouse, Title: "Senior Data Scientist", Company: "Acme",
			URL: "https://boards.greenhouse.io/acme/jobs/123", Location: "Amsterdam, NL (Hybrid)", WorkMode: "hybrid",
			SalaryMin: 75000, SalaryMax: 95000, SalaryCurrency: "EUR", SalaryPeriod: "year",
			Deadline: "2026-11-14", Posted: "2026-10-20"}, "Nice to have"},
	{"lever", "https://jobs.lever.co/globex/" + leverID + "/apply",
		"https://api.lever.co/v0/postings/globex/" + leverID + "?mode=json", "lever-globex-posting.json",
		Job{ID: "lever-board:" + leverID, Source: Lever, Title: "Biostatistician", Company: "globex",
			URL: "https://jobs.lever.co/globex/" + leverID, Location: "Remote (EU)", WorkMode: "wfh",
			SalaryMin: 5000, SalaryMax: 6000, SalaryCurrency: "EUR", SalaryPeriod: "month",
			Deadline: "2026-12-01", Posted: "2026-10-21"}, "<h3>Requirements</h3>"},
	{"json-ld", "https://careers.example.org/jobs/epi", "https://careers.example.org/jobs/epi", "page-jsonld.html",
		Job{ID: ID("https://careers.example.org/jobs/epi"), Source: Web, Title: "Epidemiologist",
			Company: "Soylent Institute", URL: "https://careers.example.org/jobs/epi", Location: "Geneva, CH",
			SalaryMin: 110000, SalaryMax: 130000, SalaryCurrency: "CHF", SalaryPeriod: "year",
			Deadline: "2026-11-20", Posted: "2026-10-19"}, "Key responsibilities"},
	{"plain page", "https://www.example.net/rs", "https://www.example.net/rs", "page-plain.html",
		Job{ID: ID("https://www.example.net/rs"), Source: Web, Title: "Research Scientist", Company: "Cyberdyne",
			URL: "https://www.example.net/rs", WorkMode: "", Deadline: "2026-12-05"}, "Run experiments."},
}

func TestFetch(t *testing.T) {
	for _, c := range fetchCases {
		t.Run(c.name, func(t *testing.T) {
			d, err := Fetch(context.Background(), newFake(t, map[string]string{c.api: c.file}), c.url)
			if err != nil {
				t.Fatal(err)
			}
			if got := withoutSummary(t, []Job{d.Job})[0]; !reflect.DeepEqual(got, c.want) {
				t.Errorf("job differs\n got: %+v\nwant: %+v", got, c.want)
			}
			if !strings.Contains(d.HTML, c.html) {
				t.Errorf("HTML lacks %q:\n%s", c.html, d.HTML)
			}
		})
	}
}

func TestFetchPlainPageDropsScriptsAndChrome(t *testing.T) {
	url := "https://www.example.net/rs"
	d, err := Fetch(context.Background(), newFake(t, map[string]string{url: "page-plain.html"}), url)
	if err != nil {
		t.Fatal(err)
	}
	text := Text(d.HTML)
	for _, bad := range []string{"tracking", "Menu", "Copyright", "color: red"} {
		if strings.Contains(text, bad) {
			t.Errorf("text contains %q: %s", bad, text)
		}
	}
}

func TestFetchErrors(t *testing.T) {
	cases := map[string]string{
		"ftp://example.org/x":           "not an http(s) URL",
		"not a url":                     "not an http(s) URL",
		"https://example.org/missing":   "HTTP 404",
		"https://boards.greenhouse.io/": "HTTP 404",
	}
	for url, want := range cases {
		_, err := Fetch(context.Background(), newFake(t, nil), url)
		if !errors.Is(err, ErrFetch) || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v, want %q", url, err, want)
		}
	}
}

func TestID(t *testing.T) {
	cases := map[string]string{
		"https://boards.greenhouse.io/acme/jobs/123":         "greenhouse-board:123",
		"https://job-boards.greenhouse.io/acme/jobs/123?x=1": "greenhouse-board:123",
		"https://jobs.eu.lever.co/globex/" + leverID:         "lever-board:" + leverID,
	}
	for url, want := range cases {
		if got := ID(url); got != want {
			t.Errorf("ID(%q) = %q, want %q", url, got, want)
		}
	}
	a, b := ID("https://example.org/a"), ID("https://example.org/b")
	if !strings.HasPrefix(a, "url:") || len(a) != 16 || a == b || a != ID("https://example.org/a") {
		t.Errorf("url IDs must be stable and distinct: %q %q", a, b)
	}
}
