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
