package source

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

const (
	adzunaURL = "https://api.adzuna.com/v1/api/jobs/nl/search/1?app_id=test-id&app_key=test-key" +
		"&content-type=application%2Fjson&results_per_page=20&what_phrase=biostatistician"
	remotiveURL   = "https://remotive.com/api/remote-jobs?search=data"
	arbeitnowURL  = "https://www.arbeitnow.com/api/job-board-api"
	greenhouseURL = "https://boards-api.greenhouse.io/v1/boards/acme/jobs?content=true"
	leverURL      = "https://api.lever.co/v0/postings/globex?mode=json"
	leverID       = "0f8e7d6c-5b4a-4321-9876-0123456789ab"
)

var adapterCases = []struct {
	name  string
	src   func(Doer) Source
	terms []string
	route [2]string
	want  []Job
}{
	{"adzuna", func(c Doer) Source {
		return &adzuna{client: c, keys: Keys{"test-id", "test-key"}, countries: []string{"nl"}}
	}, []string{"biostatistician"}, [2]string{adzunaURL, "adzuna-nl.json"}, []Job{
		{ID: "adzuna:4100000001", Source: Adzuna, Title: "Biostatistician", Company: "Initech Health",
			URL: "https://www.adzuna.nl/details/4100000001", Location: "Utrecht, Utrecht", WorkMode: "hybrid",
			SalaryMin: 60000, SalaryMax: 72000, SalaryCurrency: "EUR", SalaryPeriod: "year",
			Deadline: "2026-11-30", Posted: "2026-10-25"},
		{ID: "adzuna:4100000002", Source: Adzuna, Title: "Junior Biostatistician", Company: "Umbrella Pharma",
			URL: "https://www.adzuna.nl/details/4100000002", Location: "Leiden, Zuid-Holland", WorkMode: "wfo",
			Posted: "2026-10-24"},
	}},
	{"remotive", func(c Doer) Source { return &remotive{client: c} }, []string{"data"},
		[2]string{remotiveURL, "remotive.json"}, []Job{
			{ID: "remotive:7000001", Source: Remotive, Title: "Data Scientist", Company: "Hooli",
				URL: "https://remotive.com/remote-jobs/data/data-scientist-7000001", Location: "Europe",
				WorkMode: "wfh", SalaryMin: 80000, SalaryMax: 100000, SalaryCurrency: "USD",
				SalaryPeriod: "year", Deadline: "2026-11-30", Posted: "2026-10-26"},
			{ID: "remotive:7000002", Source: Remotive, Title: "Biostatistician", Company: "Vandelay",
				URL: "https://remotive.com/remote-jobs/data/biostatistician-7000002", Location: "Worldwide",
				WorkMode: "wfa", Posted: "2026-10-23"},
		}},
	{"arbeitnow", func(c Doer) Source { return &arbeitnow{client: c} }, []string{"data scientist", "biostatistician"},
		[2]string{arbeitnowURL, "arbeitnow.json"}, []Job{
			{ID: "arbeitnow:data-scientist-stark-berlin-12345", Source: Arbeitnow, Title: "Data Scientist (m/w/d)",
				Company: "Stark Analytics", Location: "Berlin", WorkMode: "hybrid", Posted: "2026-10-24",
				URL: "https://www.arbeitnow.com/jobs/companies/stark/data-scientist-stark-berlin-12345"},
			{ID: "arbeitnow:biostatistician-wayne-munich-23456", Source: Arbeitnow, Title: "Biostatistician",
				Company: "Wayne Research", Location: "Munich", WorkMode: "wfh", Posted: "2026-10-23",
				URL: "https://www.arbeitnow.com/jobs/companies/wayne/biostatistician-wayne-munich-23456"},
		}},
	{"greenhouse", func(c Doer) Source { return &greenhouse{client: c, boards: []string{"acme"}} },
		[]string{"data scientist"}, [2]string{greenhouseURL, "greenhouse-acme.json"}, []Job{
			{ID: "greenhouse-board:123", Source: Greenhouse, Title: "Senior Data Scientist", Company: "Acme",
				URL: "https://boards.greenhouse.io/acme/jobs/123", Location: "Amsterdam, NL (Hybrid)",
				WorkMode: "hybrid", Deadline: "2026-11-14", Posted: "2026-10-20"},
		}},
	{"lever", func(c Doer) Source { return &lever{client: c, boards: []string{"globex"}} },
		[]string{"biostatistician"}, [2]string{leverURL, "lever-globex.json"}, []Job{
			{ID: "lever-board:" + leverID, Source: Lever, Title: "Biostatistician", Company: "globex",
				URL: "https://jobs.lever.co/globex/" + leverID, Location: "Remote (EU)", WorkMode: "wfh",
				SalaryMin: 5000, SalaryMax: 6000, SalaryCurrency: "EUR", SalaryPeriod: "month",
				Deadline: "2026-12-01", Posted: "2026-10-21"},
		}},
}

// Each adapter maps a recorded response to Job, including salary, work
// mode, location and deadline when the response has them.
func TestAdaptersMapRecordedResponses(t *testing.T) {
	for _, c := range adapterCases {
		t.Run(c.name, func(t *testing.T) {
			fake := newFake(t, map[string]string{c.route[0]: c.route[1]})
			jobs, err := c.src(fake).Search(context.Background(), c.terms)
			if err != nil {
				t.Fatal(err)
			}
			if got := withoutSummary(t, jobs); !reflect.DeepEqual(got, c.want) {
				t.Errorf("jobs differ\n got: %+v\nwant: %+v", got, c.want)
			}
		})
	}
}

func TestNoTermsKeepsEverything(t *testing.T) {
	fake := newFake(t, map[string]string{"https://remotive.com/api/remote-jobs": "remotive.json",
		arbeitnowURL: "arbeitnow.json"})
	for _, s := range []Source{&remotive{client: fake}, &arbeitnow{client: fake}} {
		jobs, err := s.Search(context.Background(), nil)
		if err != nil || len(jobs) < 2 {
			t.Errorf("%s without terms: %d jobs, %v", s.Name(), len(jobs), err)
		}
	}
}

func TestAdzunaOneRequestPerCountryAndTerm(t *testing.T) {
	fake := newFake(t, nil)
	a := &adzuna{client: fake, keys: Keys{"i", "k"}, countries: []string{"nl", "DE"}}
	_, _ = a.Search(context.Background(), []string{"a", "b"})
	if len(fake.calls) != 1 || !strings.Contains(fake.calls[0], "/nl/search/1") {
		t.Errorf("first failure must stop the source: %v", fake.calls)
	}
	fake = newFake(t, map[string]string{})
	for _, u := range []string{"nl", "de"} {
		for _, term := range []string{"a", "b"} {
			fake.routes["https://api.adzuna.com/v1/api/jobs/"+u+"/search/1?app_id=i&app_key=k"+
				"&content-type=application%2Fjson&results_per_page=20&what_phrase="+term] = "adzuna-nl.json"
		}
	}
	a.client = fake
	jobs, err := a.Search(context.Background(), []string{"a", "b"})
	if err != nil || len(fake.calls) != 4 || len(jobs) != 8 {
		t.Errorf("got %d calls, %d jobs, err %v", len(fake.calls), len(jobs), err)
	}
}

func TestSearchErrorsHideAPIKeys(t *testing.T) {
	fake := newFake(t, map[string]string{adzunaURL: "!err"})
	a := &adzuna{client: fake, keys: Keys{"test-id", "test-key"}, countries: []string{"nl"}}
	_, err := a.Search(context.Background(), []string{"biostatistician"})
	if err == nil || strings.Contains(err.Error(), "test-key") || !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("got %v", err)
	}
}
