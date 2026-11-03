package source

import (
	"net/http"
	"reflect"
	"testing"
)

const (
	bcgConfig  = "../../portals/bcg.yml"
	bcgListing = "https://careers.bcg.com/widgets"
	bcgBody    = `{"lang":"en_global","deviceType":"desktop","country":"global", ` +
		`"pageName":"search-results","ddoKey":"refineSearch","from":0,"size":500, ` +
		`"jobs":true,"counts":false,"siteType":"external","keywords":"","global":true, ` +
		`"selected_fields":{}}`
)

func loadBCG(t *testing.T) Portal {
	t.Helper()
	p, err := LoadPortal(bcgConfig)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// The tracked portals/bcg.yml parses the saved Phenom widgets response: the
// untitled item is dropped, an item without location has none, and job
// pages are built from jobId. The listing has no deadline or salary.
func TestPortalBCG(t *testing.T) {
	d := &portalDoer{t: t, routes: map[string]string{bcgListing: "bcg/listing.json"}}
	got := searchPortal(t, loadBCG(t), d)
	co, src, page := "Boston Consulting Group", "portal:bcg", "https://careers.bcg.com/global/en/job/"
	want := []Job{
		{ID: src + ":900201", Source: src, Title: "Consultant, Fictional Public Sector Practice", Company: co,
			Location: "Berlin, Berlin, Germany", URL: page + "900201"},
		{ID: src + ":900202", Source: src, Title: "Project Leader, Example Global Health", Company: co,
			Location: "Nairobi, Nairobi County, Kenya", URL: page + "900202"},
		{ID: src + ":900203", Source: src, Title: "Data Scientist, Imaginary Analytics Lab", Company: co,
			URL: page + "900203"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
	r := d.reqs[0]
	if r.Method != http.MethodPost || d.bodies[0] != bcgBody ||
		r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Accept") != "application/json" {
		t.Errorf("request: %s %v %q", r.Method, r.Header, d.bodies[0])
	}
}

// A BCG job page is recognised as the bcg portal, with its job ID and ats;
// the slugged form the site links to is not (see the config's notes).
func TestPortalBCGMatches(t *testing.T) {
	p := loadBCG(t)
	if got := p.Hosts(); !reflect.DeepEqual(got, []string{"careers.bcg.com"}) {
		t.Errorf("hosts = %v", got)
	}
	cases := map[string]string{
		"https://careers.bcg.com/global/en/job/900201":                 "portal:bcg:900201",
		"https://careers.bcg.com/global/en/job/900201#apply":           "portal:bcg:900201",
		"https://careers.bcg.com/global/en/job/900201/Consultant-Slug": "",
		"https://careers.bcg.com/global/en/search-results":             "",
		"https://experiencedtalent.bcg.com/global/en/job/900201":       "",
		"https://careers.bcg.com/jp/ja/job/900201":                     "",
	}
	for u, want := range cases {
		got, id, ok := PortalFor([]Portal{p}, u)
		if ok != (want != "") || id != want || ok && (got.Org != "bcg" || got.ATS != "phenom") {
			t.Errorf("%s: %q %v %q, want %q", u, got.Org, ok, id, want)
		}
	}
}
