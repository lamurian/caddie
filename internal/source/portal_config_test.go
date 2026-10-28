package source

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
)

const portalDir = "../../testdata/portals"

func TestLoadPortalInvalid(t *testing.T) {
	cases := map[string]string{ // file in invalid/ → field the error names
		"missing-name":           "name: required",
		"missing-listing-url":    "listing_url: required",
		"relative-listing-url":   "listing_url: must be an absolute",
		"missing-ats":            "ats: required",
		"bad-method":             `method: "PUT" is not one of GET|POST`,
		"body-without-post":      "body: only allowed with method: POST",
		"both-selector-and-json": "item_selector, json_path: set exactly one, not both",
		"no-selector-or-json":    "item_selector, json_path: set exactly one",
		"missing-title":          "fields.title: required",
		"missing-url":            "fields.url: required unless",
		"detail-url-no-id":       "detail_url: must contain {id}",
		"unknown-key":            "field selector not found",
	}
	for name, want := range cases {
		path := filepath.Join(portalDir, "invalid", name+".yml")
		_, err := LoadPortal(path)
		if !errors.Is(err, config.ErrConfig) {
			t.Errorf("%s: got %v, want ErrConfig", name, err)
			continue
		}
		if msg := err.Error(); !strings.Contains(msg, path+": ") || !strings.Contains(msg, want) {
			t.Errorf("%s: got %q, want the file and %q", name, msg, want)
		}
	}
}

func TestLoadPortalExamples(t *testing.T) {
	html, err := LoadPortal(filepath.Join(portalDir, "example-html.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if html.Org != "example-html" || html.Method != "GET" || html.Fields.URL != "a.title@href" ||
		html.SourceName() != "portal:example-html" || html.ATS != "taleo" {
		t.Errorf("example-html: got %+v", html)
	}
	js, err := LoadPortal(filepath.Join(portalDir, "example-json.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if js.Method != "POST" || js.JSONPath != "data.results" || js.Fields.Location != "location.city" {
		t.Errorf("example-json: got %+v", js)
	}
}

func TestLoadPortalJSRendered(t *testing.T) {
	path := filepath.Join(t.TempDir(), "spa.yml")
	data := "name: Example\nlisting_url: https://careers.example.org/\nats: workday\njs_rendered: true\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if p, err := LoadPortal(path); err != nil || !p.JSRendered {
		t.Errorf("js_rendered without selectors: got %+v, %v", p, err)
	}
}

func TestLoadPortals(t *testing.T) {
	got, notices, err := LoadPortals(portalDir, []string{"example-json", "bain", "example-html"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Org != "example-json" || got[1].Org != "example-html" {
		t.Errorf("portals: got %+v", got)
	}
	want := []string{"portal:bain: skipped: " + filepath.Join(portalDir, "bain.yml") + " does not exist"}
	if !reflect.DeepEqual(notices, want) {
		t.Errorf("notices: got %q, want %q", notices, want)
	}
	for _, orgs := range [][]string{{"../invalid/unknown-key"}, {"Who"}} {
		if _, _, err := LoadPortals(portalDir, orgs); !errors.Is(err, config.ErrConfig) {
			t.Errorf("%v: got %v, want ErrConfig", orgs, err)
		}
	}
	if _, _, err := LoadPortals(filepath.Join(portalDir, "invalid"), []string{"missing-ats"}); !errors.Is(err, config.ErrConfig) {
		t.Errorf("invalid config: got %v, want ErrConfig", err)
	}
}

func TestPortalHosts(t *testing.T) {
	p := Portal{ListingURL: "https://Jobs.example.net/api/search", DetailURL: "https://careers.example.net/job/{id}"}
	if got := p.Hosts(); !reflect.DeepEqual(got, []string{"careers.example.net", "jobs.example.net"}) {
		t.Errorf("two hosts: got %q", got)
	}
	p.DetailURL = "https://jobs.example.net/job/{id}"
	if got := p.Hosts(); !reflect.DeepEqual(got, []string{"jobs.example.net"}) {
		t.Errorf("one host: got %q", got)
	}
}

func TestPortalMatches(t *testing.T) {
	query := Portal{ListingURL: "https://careers.example.org/vacancies",
		DetailURL: "https://careers.example.org/job?id={id}"}
	path := Portal{ListingURL: "https://jobs.example.net/api/search",
		DetailURL: "https://jobs.example.net/careers/job/{id}"}
	none := Portal{ListingURL: "https://careers.example.org/vacancies"}
	cases := []struct {
		p      Portal
		url    string
		id     string
		wantOK bool
	}{
		{query, "https://careers.example.org/job?id=4821", "4821", true},
		{query, "https://CAREERS.example.org/job?id=4821#apply", "4821", true},
		{query, "https://careers.example.org/job?id=4821&lang=en", "", false},
		{query, "https://careers.example.org/jobs?id=4821", "", false},
		{query, "https://evil.example.com/job?id=4821", "", false},
		{path, "https://jobs.example.net/careers/job/JR-77", "JR-77", true},
		{path, "https://jobs.example.net/careers/job/JR-77/apply", "", false},
		{path, "https://jobs.example.net/careers/job/", "", false},
		{none, "https://careers.example.org/anything/12", "", true},
		{none, "https://other.example.org/anything/12", "", false},
		{none, "not a url", "", false},
	}
	for _, c := range cases {
		id, ok := c.p.Matches(c.url)
		if id != c.id || ok != c.wantOK {
			t.Errorf("%s: got (%q, %v), want (%q, %v)", c.url, id, ok, c.id, c.wantOK)
		}
	}
}
