package source

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
)

func names(srcs []Source) []string {
	var out []string
	for _, s := range srcs {
		out = append(out, s.Name())
	}
	return out
}

func allEnabled() config.Sources {
	return config.Sources{
		Adzuna:   config.Adzuna{Enabled: true, Countries: []string{"nl"}},
		Remotive: config.Toggle{Enabled: true}, Arbeitnow: config.Toggle{Enabled: true},
		GreenhouseBoards: []string{"acme"}, LeverBoards: []string{"globex"},
	}
}

func TestBuildEnabledSources(t *testing.T) {
	srcs, notices := Build(allEnabled(), Keys{"id", "key"}, "", nil)
	if !reflect.DeepEqual(names(srcs), Names) || len(notices) != 0 {
		t.Errorf("all enabled: %v, notices %v", names(srcs), notices)
	}
	srcs, notices = Build(allEnabled(), Keys{"id", "key"}, Lever, nil)
	if !reflect.DeepEqual(names(srcs), []string{Lever}) || len(notices) != 0 {
		t.Errorf("--source lever-board: %v, notices %v", names(srcs), notices)
	}
}

// A disabled source, or missing Adzuna keys, is skipped with a notice.
func TestBuildSkipsWithNotice(t *testing.T) {
	cases := []struct {
		name   string
		prefs  func(*config.Sources)
		keys   Keys
		only   string
		notice string
	}{
		{"adzuna keys missing", func(*config.Sources) {}, Keys{}, "", "adzuna: skipped: CADDIE_ADZUNA_APP_ID"},
		{"adzuna key half set", func(*config.Sources) {}, Keys{AdzunaAppID: "id"}, "", "CADDIE_ADZUNA_APP_KEY"},
		{"adzuna disabled", func(s *config.Sources) { s.Adzuna.Enabled = false }, Keys{"i", "k"}, "",
			"adzuna: skipped: disabled in preferences.yml"},
		{"adzuna no countries", func(s *config.Sources) { s.Adzuna.Countries = nil }, Keys{"i", "k"}, "",
			"sources.adzuna.countries"},
		{"remotive disabled", func(s *config.Sources) { s.Remotive.Enabled = false }, Keys{"i", "k"}, "",
			"remotive: skipped: disabled"},
		{"arbeitnow only and disabled", func(s *config.Sources) { s.Arbeitnow.Enabled = false }, Keys{"i", "k"},
			Arbeitnow, "arbeitnow: skipped: disabled"},
		{"no greenhouse boards", func(s *config.Sources) { s.GreenhouseBoards = nil }, Keys{"i", "k"}, "",
			"greenhouse-board: skipped: sources.greenhouse_boards is empty"},
		{"no lever boards", func(s *config.Sources) { s.LeverBoards = nil }, Keys{"i", "k"}, "",
			"lever-board: skipped: sources.lever_boards is empty"},
	}
	for _, c := range cases {
		prefs := allEnabled()
		c.prefs(&prefs)
		srcs, notices := Build(prefs, c.keys, c.only, nil)
		if len(notices) != 1 || !strings.Contains(notices[0], c.notice) {
			t.Errorf("%s: notices %v, want %q", c.name, notices, c.notice)
		}
		want := len(Names) - 1
		if c.only != "" {
			want = 0
		}
		if len(srcs) != want {
			t.Errorf("%s: %d sources, want %d", c.name, len(srcs), want)
		}
	}
}

type stubSource struct {
	name string
	jobs []Job
	err  error
}

func (s stubSource) Name() string { return s.name }
func (s stubSource) Search(context.Context, []string) ([]Job, error) {
	return s.jobs, s.err
}

func job(id, url, title, company string) Job {
	return Job{ID: id, URL: url, Title: title, Company: company}
}

func ids(jobs []Job) []string {
	out := []string{}
	for _, j := range jobs {
		out = append(out, j.ID)
	}
	return out
}

var stubs = []Source{
	stubSource{name: "a", jobs: []Job{job("a:1", "u1", "Data Scientist", "Acme"),
		job("a:2", "u2", "Sales Engineer", "Acme"), job("a:3", "u3", "Biostatistician", "ExampleCorp")}},
	stubSource{name: "b", jobs: []Job{job("b:1", "u1", "Data Scientist", "Acme"),
		job("b:2", "u4", "Epidemiologist", "Globex")}},
	stubSource{name: "c", err: errors.New("fetch failed: HTTP 500")},
}

func TestSearchFiltersAndDeduplicates(t *testing.T) {
	f := Filter{ExcludeRoles: []string{"sales engineer"}, ExcludeCompanies: []string{" examplecorp"}}
	res, err := Search(context.Background(), stubs, nil, f)
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(res.Jobs); !reflect.DeepEqual(got, []string{"a:1", "b:2"}) {
		t.Errorf("jobs = %v (URL u1 must appear once, excluded role and company dropped)", got)
	}
	if len(res.Notices) != 1 || !strings.HasPrefix(res.Notices[0], "c: skipped: ") {
		t.Errorf("notices = %v", res.Notices)
	}
}

// Seen IDs are hidden unless --include-seen leaves Filter.Seen nil.
func TestSearchSeenFiltering(t *testing.T) {
	seen := map[string]bool{"a:1": true, "zz:9": true}
	res, _ := Search(context.Background(), stubs, nil, Filter{Seen: seen})
	if got := ids(res.Jobs); !reflect.DeepEqual(got, []string{"a:2", "a:3", "b:2"}) || res.HiddenSeen != 1 {
		t.Errorf("excluding seen: %v, hidden %d", got, res.HiddenSeen)
	}
	res, _ = Search(context.Background(), stubs, nil, Filter{})
	if got := ids(res.Jobs); len(got) != 4 || res.HiddenSeen != 0 {
		t.Errorf("including seen: %v, hidden %d", got, res.HiddenSeen)
	}
	res, _ = Search(context.Background(), stubs, nil, Filter{Limit: 2})
	if len(res.Jobs) != 2 {
		t.Errorf("limit: %v", ids(res.Jobs))
	}
}

func TestSearchFailsOnlyWhenEverySourceFails(t *testing.T) {
	_, err := Search(context.Background(), stubs[2:], nil, Filter{})
	if !errors.Is(err, ErrFetch) {
		t.Errorf("got %v, want ErrFetch", err)
	}
	res, err := Search(context.Background(), nil, nil, Filter{})
	if err != nil || res.Jobs == nil || res.Notices == nil {
		t.Errorf("no sources: %+v, %v (empty lists, not null)", res, err)
	}
}

func TestKnown(t *testing.T) {
	if !Known(Greenhouse) || Known("portal") || Known("") {
		t.Error("Known is wrong")
	}
}
