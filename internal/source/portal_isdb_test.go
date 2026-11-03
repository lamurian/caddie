package source

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

const (
	isdbConfig  = "../../portals/isdb.yml"
	isdbListing = "https://careers.isdb.org/search/?q=&locale=en_GB"
	isdbJob     = "https://careers.isdb.org/job/"
)

func loadIsDB(t *testing.T) Portal {
	t.Helper()
	p, err := LoadPortal(isdbConfig)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// The tracked portals/isdb.yml parses the saved SuccessFactors listing:
// titles come from the desktop section only (each tile repeats them), the
// relative data-url resolves against the listing, a tile without a
// Location field has no location, and IDs are URL hashes.
func TestPortalIsDB(t *testing.T) {
	p := loadIsDB(t)
	d := &portalDoer{t: t, routes: map[string]string{isdbListing: "isdb/listing.html"}}
	got := searchPortal(t, p, d)
	co, src := "Islamic Development Bank", "portal:isdb"
	job := func(path, title, loc string) Job {
		u := isdbJob + path
		return Job{ID: PortalJobID(p, "", u), Source: src, Title: title, Company: co, Location: loc, URL: u}
	}
	want := []Job{
		job("Senior-Economist%2C-Fictional-Finance-%28ICD%29/9900101155/",
			"Senior Economist, Fictional Finance (ICD)", "SA"),
		job("Programme-Officer-Regional-Hub-Atlantis/9900202155/",
			"Programme Officer - Regional Hub Atlantis", "ID"),
		job("Data-Analyst-%28Outsourced%29/9900303155/", "Data Analyst (Outsourced)", ""),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
	for _, j := range got {
		if !strings.HasPrefix(j.ID, src+":") || len(j.ID) != len(src)+13 {
			t.Errorf("id %q is not a 12-hex URL hash", j.ID)
		}
	}
	if r := d.reqs[0]; len(d.reqs) != 1 || r.Method != http.MethodGet || d.bodies[0] != "" {
		t.Errorf("requests: %d, %s %q", len(d.reqs), r.Method, d.bodies[0])
	}
}

// Without a detail_url any careers.isdb.org URL is the isdb portal, and a
// job page gets the same hash ID that search gives it.
func TestPortalIsDBMatches(t *testing.T) {
	p := loadIsDB(t)
	if got := p.Hosts(); !reflect.DeepEqual(got, []string{"careers.isdb.org"}) {
		t.Errorf("hosts = %v", got)
	}
	job := isdbJob + "Data-Analyst-%28Outsourced%29/9900303155/"
	cases := map[string]bool{
		job:                                  true,
		"https://careers.isdb.org/search/":   true,
		"https://www.isdb.org/careers":       false,
		"https://career55.sapsf.eu/career?x": false,
	}
	for u, want := range cases {
		got, id, ok := PortalFor([]Portal{p}, u)
		if ok != want || ok && (got.Org != "isdb" || got.ATS != "successfactors" || id != PortalJobID(p, "", u)) {
			t.Errorf("%s: %q %v %q, want %v", u, got.Org, ok, id, want)
		}
	}
}
