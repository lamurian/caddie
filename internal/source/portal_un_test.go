package source

import (
	"net/http"
	"reflect"
	"testing"
)

const (
	unConfig  = "../../portals/un.yml"
	unListing = "https://careers.un.org/api/public/opening/jo/list/filteredV2/en"
	unBody    = `{"filterConfig":{"keyword":""}, ` +
		`"pagination":{"page":0,"itemPerPage":500,"sortBy":"startDate","sortDirection":-1}}`
)

func loadUN(t *testing.T) Portal {
	t.Helper()
	p, err := LoadPortal(unConfig)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// The tracked portals/un.yml parses the saved UN listing: the untitled item
// is dropped, a null endDate gives no deadline, and the UTC endDate gives its
// own date. Location is the first duty station only.
func TestPortalUN(t *testing.T) {
	d := &portalDoer{t: t, routes: map[string]string{unListing: "un/listing.json"}}
	got := searchPortal(t, loadUN(t), d)
	co, src, page := "United Nations", "portal:un", "https://careers.un.org/jobSearchDescription/"
	want := []Job{
		{ID: src + ":900101", Source: src, Title: "Programme Management Officer, Fictional Affairs", Company: co,
			Location: "Nairobi",
			URL:      page + "900101?language=en", Deadline: "2026-11-28"},
		{ID: src + ":900102", Source: src, Title: "Statistician (Temporary Job Opening)", Company: co,
			Location: "New York",
			URL:      page + "900102?language=en", Deadline: "2026-12-15"},
		{ID: src + ":900103", Source: src, Title: "Associate Economic Affairs Officer", Company: co,
			Location: "Bangkok",
			URL:      page + "900103?language=en"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
	r := d.reqs[0]
	if r.Method != http.MethodPost || d.bodies[0] != unBody ||
		r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Accept") != "application/json" {
		t.Errorf("request: %s %v %q", r.Method, r.Header, d.bodies[0])
	}
}

// A UN job page is recognised as the un portal, with its job ID and ats.
func TestPortalUNMatches(t *testing.T) {
	p := loadUN(t)
	if got := p.Hosts(); !reflect.DeepEqual(got, []string{"careers.un.org"}) {
		t.Errorf("hosts = %v", got)
	}
	cases := map[string]string{
		"https://careers.un.org/jobSearchDescription/900101?language=en":       "portal:un:900101",
		"https://careers.un.org/jobSearchDescription/900101?language=en#apply": "portal:un:900101",
		"https://careers.un.org/jobSearchDescription/900101?language=fr":       "",
		"https://careers.un.org/home?language=en":                              "",
		"https://inspira.un.org/jobSearchDescription/900101?language=en":       "",
	}
	for u, want := range cases {
		got, id, ok := PortalFor([]Portal{p}, u)
		if ok != (want != "") || id != want || ok && (got.Org != "un" || got.ATS != "inspira") {
			t.Errorf("%s: %q %v %q, want %q", u, got.Org, ok, id, want)
		}
	}
}
