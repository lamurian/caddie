package source

import (
	"net/http"
	"reflect"
	"testing"
)

const (
	whoConfig  = "../../portals/who.yml"
	whoListing = "https://careers.who.int/careersection/rest/jobboard/searchjobs?lang=en&portal=101430233"
	whoBody    = `{"multilineEnabled":false, ` +
		`"sortingSelection":{"sortBySelectionParam":"3","ascendingSortingOrder":"false"}, ` +
		`"fieldData":{"fields":{"KEYWORD":"","LOCATION":""},"valid":true}, ` +
		`"filterSelectionParam":{"searchFilterSelections":[]}, ` +
		`"advancedSearchFiltersSelectionParam":{"searchFilterSelections":[]},"pageNo":1}`
)

func loadWHO(t *testing.T) Portal {
	t.Helper()
	p, err := LoadPortal(whoConfig)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// The tracked portals/who.yml parses the saved Taleo listing: title and
// location come from the positional column array, the location's JSON-array
// string is joined, an untitled item is dropped and an empty location ("[]")
// stays empty. The listing has no deadlines (column 2 is the posting date).
func TestPortalWHO(t *testing.T) {
	d := &portalDoer{t: t, routes: map[string]string{whoListing: "who/listing.json"}}
	got := searchPortal(t, loadWHO(t), d)
	co, src := "World Health Organization", "portal:who"
	page := "https://careers.who.int/careersection/ex/jobdetail.ftl?job="
	want := []Job{
		{ID: src + ":2690101", Source: src, Title: "Technical Officer, Fictional Disease Surveillance",
			Company: co, URL: page + "2690101&lang=en", Location: "Switzerland-Geneva"},
		{ID: src + ":2690102", Source: src, Title: "National Consultant : Santé publique imaginaire",
			Company: co, URL: page + "2690102&lang=en", Location: "Burkina Faso-Ouagadougou, Mali-Bamako"},
		{ID: src + ":2690104", Source: src, Title: "Budget and Finance Officer (BFO)",
			Company: co, URL: page + "2690104&lang=en"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
	r := d.reqs[0]
	if r.Method != http.MethodPost || d.bodies[0] != whoBody || r.Header.Get("tz") != "GMT+00:00" ||
		r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Accept") != "application/json" {
		t.Errorf("request: %s %v %q", r.Method, r.Header, d.bodies[0])
	}
}

// A WHO job page is recognised as the who portal, with its job ID and ats.
func TestPortalWHOMatches(t *testing.T) {
	p := loadWHO(t)
	if got := p.Hosts(); !reflect.DeepEqual(got, []string{"careers.who.int"}) {
		t.Errorf("hosts = %v", got)
	}
	cases := map[string]string{
		"https://careers.who.int/careersection/ex/jobdetail.ftl?job=2690101&lang=en":     "portal:who:2690101",
		"https://careers.who.int/careersection/ex/jobdetail.ftl?job=2690101&lang=en#top": "portal:who:2690101",
		"https://careers.who.int/careersection/ex/jobdetail.ftl?job=2690101&lang=fr":     "",
		"https://careers.who.int/careersection/ex/jobsearch.ftl?lang=en":                 "",
		"https://www.who.int/careersection/ex/jobdetail.ftl?job=2690101&lang=en":         "",
	}
	for u, want := range cases {
		got, id, ok := PortalFor([]Portal{p}, u)
		if ok != (want != "") || id != want || ok && (got.Org != "who" || got.ATS != "taleo") {
			t.Errorf("%s: %q %v %q, want %q", u, got.Org, ok, id, want)
		}
	}
}
