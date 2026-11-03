package source

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
)

func TestLoadAllPortals(t *testing.T) {
	portals, err := LoadAllPortals(portalDir)
	if err != nil {
		t.Fatal(err)
	}
	var orgs []string
	for _, p := range portals {
		orgs = append(orgs, p.Org)
	}
	if !slices.IsSorted(orgs) || !slices.Contains(orgs, "example-html") || !slices.Contains(orgs, "example-json") {
		t.Errorf("orgs = %v", orgs)
	}
	if got, err := LoadAllPortals(filepath.Join(t.TempDir(), "absent")); got != nil || err != nil {
		t.Errorf("missing dir: %v, %v", got, err)
	}
	invalid := filepath.Join(portalDir, "invalid")
	if _, err := LoadAllPortals(invalid); !errors.Is(err, config.ErrConfig) || !strings.Contains(err.Error(), invalid) {
		t.Errorf("invalid config: got %v", err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Bad_Name.yml"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAllPortals(dir); !errors.Is(err, config.ErrConfig) || !strings.Contains(err.Error(), "Bad_Name.yml") {
		t.Errorf("misnamed file: got %v", err)
	}
}

func TestPortalFor(t *testing.T) {
	html, err := LoadPortal(filepath.Join(portalDir, "example-html.yml"))
	if err != nil {
		t.Fatal(err)
	}
	js, err := LoadPortal(filepath.Join(portalDir, "example-json.yml"))
	if err != nil {
		t.Fatal(err)
	}
	noDetail := html
	noDetail.Org, noDetail.DetailURL = "no-detail", ""
	portals := []Portal{js, html}
	cases := []struct{ url, org, id string }{
		{"https://careers.example.org/job?id=1001", "example-html", "portal:example-html:1001"},
		{"https://jobs.example.net/careers/job/J-7", "example-json", "portal:example-json:J-7"},
		{"https://careers.example.org/about", "", ""},
		{"https://www.example.net/rs", "", ""},
	}
	for _, c := range cases {
		p, id, ok := PortalFor(portals, c.url)
		if ok != (c.org != "") || p.Org != c.org || id != c.id {
			t.Errorf("%s: got %q %q %v", c.url, p.Org, id, ok)
		}
	}
	// Without detail_url any listing-host page matches, with the URL hash.
	u := "https://careers.example.org/about"
	_, id, ok := PortalFor([]Portal{noDetail}, u)
	if want := "portal:no-detail:" + strings.TrimPrefix(ID(u), "url:"); !ok || id != want {
		t.Errorf("no detail_url: got %q %v, want %q", id, ok, want)
	}
}

func TestSourceOf(t *testing.T) {
	cases := map[string]string{
		"portal:who:123":            "portal:who",
		"portal:example-html:1:2":   "portal:example-html",
		"portal:who":                "",
		"portal:Bad_Org:1":          "",
		"adzuna:42":                 Adzuna,
		"greenhouse-board:123":      Greenhouse,
		"url:0123456789ab":          "",
		"nonsense":                  "",
		"somewhere-else:portal:x:1": "",
	}
	for id, want := range cases {
		if got := SourceOf(id); got != want {
			t.Errorf("SourceOf(%q) = %q, want %q", id, got, want)
		}
	}
}

// A trailing /{slug} matches the bare job page and one extra segment, and
// is dropped when building a job URL from an id (ADR 0031).
func TestPortalDetailSlug(t *testing.T) {
	p := Portal{File: "x.yml", Name: "Example", ListingURL: "https://jobs.example.net/api", Method: "GET", ATS: "bespoke",
		JSONPath: "jobs", Fields: PortalFields{ID: "id", Title: "t"},
		DetailURL: "https://jobs.example.net/en/job/{id}/{slug}"}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"https://jobs.example.net/en/job/J-7":              "J-7",
		"https://jobs.example.net/en/job/J-7/Analyst-Role": "J-7",
		"https://jobs.example.net/en/job/J-7/Analyst#top":  "J-7",
		"https://jobs.example.net/en/job/J-7/a/b":          "",
		"https://jobs.example.net/en/job/J-7/a?x=1":        "",
		"https://jobs.example.net/fr/job/J-7/a":            "",
	}
	for u, want := range cases {
		if id, ok := p.Matches(u); id != want || ok != (want != "") {
			t.Errorf("%s: got (%q, %v), want %q", u, id, ok, want)
		}
	}
	if got := p.detailFor("J 7"); got != "https://jobs.example.net/en/job/J%207" {
		t.Errorf("detailFor = %q", got)
	}
	for _, bad := range []string{"https://jobs.example.net/job/{id}/{slug}?x=1",
		"https://jobs.example.net/job/{slug}/{id}/{slug}", "https://jobs.example.net/job/{id}-{slug}"} {
		p.DetailURL = bad
		if err := p.Validate(); !errors.Is(err, config.ErrConfig) || !strings.Contains(err.Error(), "{slug}") {
			t.Errorf("%s: got %v", bad, err)
		}
	}
}
