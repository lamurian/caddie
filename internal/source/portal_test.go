package source

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// portalDoer serves files from testdata/portals/ (or inline bodies starting
// with "<") by URL and records each request's method, headers and body.
type portalDoer struct {
	t      *testing.T
	routes map[string]string
	reqs   []*http.Request
	bodies []string
}

func (d *portalDoer) Do(req *http.Request) (*http.Response, error) {
	d.reqs = append(d.reqs, req)
	body := ""
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		body = string(b)
	}
	d.bodies = append(d.bodies, body)
	src, ok := d.routes[req.URL.String()]
	switch {
	case !ok:
		return nil, errors.New("connection refused")
	case strings.HasPrefix(src, "<"):
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(src))}, nil
	}
	data, err := os.ReadFile(filepath.Join(portalDir, src))
	if err != nil {
		d.t.Fatal(err)
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(data))}, nil
}

func loadExample(t *testing.T, org string) Portal {
	t.Helper()
	p, err := LoadPortal(filepath.Join(portalDir, org+".yml"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func searchPortal(t *testing.T, p Portal, d *portalDoer, terms ...string) []Job {
	t.Helper()
	jobs, err := (&portalSource{p: p, client: d}).Search(context.Background(), terms)
	if err != nil {
		t.Fatal(err)
	}
	return jobs
}

func TestPortalSearchHTML(t *testing.T) {
	d := &portalDoer{t: t, routes: map[string]string{
		"https://careers.example.org/vacancies": "example-html/listing.html"}}
	got := searchPortal(t, loadExample(t, "example-html"), d)
	co, src := "Example Health Organization", "portal:example-html"
	want := []Job{
		{ID: src + ":EHO-2041", Source: src, Title: "Technical Officer, Health Data Systems", Company: co,
			URL: "https://careers.example.org/job?id=EHO-2041", Location: "Geneva, Switzerland", Deadline: "2026-11-30"},
		{ID: src + ":EHO-2057", Source: src, Title: "Epidemiologist", Company: co,
			URL: "https://careers.example.org/job?id=EHO-2057", Location: "Remote", WorkMode: "wfh", Deadline: "2026-12-15"},
		{ID: src + ":EHO-2063", Source: src, Title: "Programme Assistant (Fixed-term)", Company: co,
			URL: "https://careers.example.org/job?id=EHO-2063", Deadline: "2027-01-08"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
	r := d.reqs[0]
	if r.Method != http.MethodGet || r.Header.Get("Accept") != "text/html" || d.bodies[0] != "" {
		t.Errorf("request: %s %v %q", r.Method, r.Header, d.bodies[0])
	}
	if got := searchPortal(t, loadExample(t, "example-html"), d, "EPIDEMIOLOGIST", "nurse"); len(got) != 1 ||
		got[0].Title != "Epidemiologist" {
		t.Errorf("term filter: %+v", got)
	}
}

func TestPortalSearchJSON(t *testing.T) {
	d := &portalDoer{t: t, routes: map[string]string{
		"https://jobs.example.net/api/search": "example-json/listing.json"}}
	got := searchPortal(t, loadExample(t, "example-json"), d)
	co, src := "Example Development Bank", "portal:example-json"
	want := []Job{
		{ID: src + ":9007199254740993123", Source: src, Title: "Senior Economist, Climate Finance", Company: co,
			URL: "https://jobs.example.net/careers/job/9007199254740993123", Location: "Manila",
			SalaryMin: 120000, SalaryMax: 150000, SalaryCurrency: "USD", SalaryPeriod: "year", Deadline: "2026-11-30"},
		{ID: src + ":4502", Source: src, Title: "Procurement Officer", Company: co,
			URL: "https://jobs.example.net/careers/job/4502", Location: "Abidjan, Nairobi", Deadline: "2026-12-15"},
		{ID: src + ":4504", Source: src, Title: "Data Analyst (Consultant)", Company: co,
			URL: "https://jobs.example.net/careers/job/4504?lang=en"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
	r := d.reqs[0]
	if r.Method != http.MethodPost || d.bodies[0] != `{"page": 1, "size": 100}` ||
		r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Accept") != "application/json" {
		t.Errorf("request: %s %v %q", r.Method, r.Header, d.bodies[0])
	}
}

// Without fields.id the native ID is detail_url's {id}, else a URL hash;
// an id without url fills detail_url, escaped for its query or path.
func TestPortalIDsAndURLs(t *testing.T) {
	page := `<li><a href="/job?id=A1">Analyst</a></li><li><a href="/news/x">Clerk</a></li><li>No link</li>`
	p := Portal{Org: "x", Name: "X", ListingURL: "https://careers.example.org/list", ItemSelector: "li",
		Fields: PortalFields{Title: "li", URL: "a@href"}, DetailURL: "https://careers.example.org/job?id={id}"}
	d := &portalDoer{t: t, routes: map[string]string{p.ListingURL: page}}
	got := ids(searchPortal(t, p, d))
	if want := []string{"portal:x:A1", "portal:x:" + strings.TrimPrefix(ID("https://careers.example.org/news/x"), "url:")}; !reflect.DeepEqual(got, want) {
		t.Errorf("ids = %v, want %v", got, want)
	}
	p.DetailURL = "https://careers.example.org/job?id={id}"
	if u := p.detailFor("A 1/2"); u != "https://careers.example.org/job?id=A+1%2F2" {
		t.Errorf("query {id}: %s", u)
	}
	p.DetailURL = "https://careers.example.org/jobs/{id}/view"
	if u := p.detailFor("A 1/2"); u != "https://careers.example.org/jobs/A%201%2F2/view" {
		t.Errorf("path {id}: %s", u)
	}
}

func TestPortalSalaryNeedsCurrencyOrPeriod(t *testing.T) {
	p := Portal{Org: "x", Name: "X"}
	for in, want := range map[string]int{"P-3": 0, "G-5": 0, "CHF 90,000": 90000, "120,000 - 150,000": 120000} {
		j, _ := p.job(portalItem{"title": "T", "url": "https://x.example/1", "salary": in})
		if j.SalaryMin != want {
			t.Errorf("salary %q: min %d, want %d", in, j.SalaryMin, want)
		}
	}
}

func TestBuildPortals(t *testing.T) {
	html, js := loadExample(t, "example-html"), loadExample(t, "example-json")
	jsr := Portal{Org: "un", Name: "UN", JSRendered: true}
	prefs := allEnabled()
	prefs.Portals = []string{"example-html", "example-json", "un"}
	portals := []Portal{html, js, jsr}
	cases := []struct {
		only, notice string
		want         []string
	}{
		{"", "portal:un: skipped: listing needs JavaScript (see portals/un.yml notes)",
			append(append([]string{}, Names...), "portal:example-html", "portal:example-json")},
		{"portal", "portal:un: skipped: listing needs JavaScript", []string{"portal:example-html", "portal:example-json"}},
		{"portal:example-json", "", []string{"portal:example-json"}},
		{"portal:who", "portal:who: skipped: not in sources.portals in preferences.yml", nil},
		{Remotive, "", []string{Remotive}},
	}
	for _, c := range cases {
		srcs, notices := Build(prefs, Keys{"id", "key"}, c.only, nil, portals)
		if !reflect.DeepEqual(names(srcs), c.want) || !strings.Contains(strings.Join(notices, "\n"), c.notice) ||
			(c.notice == "") != (len(notices) == 0) {
			t.Errorf("--source %q: %v, notices %v", c.only, names(srcs), notices)
		}
	}
	prefs.Portals = nil
	if _, notices := Build(prefs, Keys{"id", "key"}, "portal", nil, nil); len(notices) != 1 ||
		!strings.Contains(notices[0], "sources.portals is empty") {
		t.Errorf("no portals enabled: %v", notices)
	}
}

func TestPortalOrgsAndKnown(t *testing.T) {
	enabled := []string{"who", "un"}
	for only, want := range map[string][]string{"": enabled, "portal": enabled, "portal:un": {"un"}, "lever-board": nil} {
		if got := PortalOrgs(enabled, only); !reflect.DeepEqual(got, want) {
			t.Errorf("PortalOrgs(%q) = %v, want %v", only, got, want)
		}
	}
	for name, want := range map[string]bool{"portal": true, "portal:who": true, "portal:un-dpo2": true,
		"portal:": false, "portal:WHO": false, "portals": false, "remotive": true, "bogus": false} {
		if Known(name) != want {
			t.Errorf("Known(%q) = %v", name, !want)
		}
	}
}

// A failing portal is a notice; the other sources still return.
func TestPortalFailureIsNotice(t *testing.T) {
	d := &portalDoer{t: t, routes: map[string]string{}}
	srcs := []Source{&portalSource{p: loadExample(t, "example-html"), client: d},
		stubSource{name: Remotive, jobs: []Job{job("remotive:1", "https://r.example/1", "Analyst", "R")}}}
	res, err := Search(context.Background(), srcs, nil, Filter{})
	if err != nil || len(res.Jobs) != 1 || len(res.Notices) != 1 ||
		!strings.HasPrefix(res.Notices[0], "portal:example-html: skipped: fetch failed: https://careers.example.org/vacancies") {
		t.Errorf("got %+v, %v", res, err)
	}
}
