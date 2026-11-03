package source

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	bcgConfig  = "../../portals/bcg.yml"
	bcgListing = "https://careers.bcg.com/widgets"
	bcgBody    = `{"lang":"en_global","deviceType":"desktop","country":"global", ` +
		`"pageName":"search-results","ddoKey":"refineSearch","from":{offset},"size":500, ` +
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

// bcgPage is the request key for offset from, as pageDoer sees it.
func bcgPage(from string) string {
	return bcgListing + " " + strings.ReplaceAll(bcgBody, "{offset}", from)
}

func bcgFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(portalDir, "bcg", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// The tracked portals/bcg.yml parses the saved Phenom widgets responses: the
// untitled item is dropped, an item without location has none, and job
// pages are built from jobId. It pages by "from" in steps of 500 (ADR
// 0030), drops an item repeated on page 2, and stops at the empty page at
// 1000. The listing has no deadline or salary.
func TestPortalBCG(t *testing.T) {
	d := &pageDoer{pages: map[string]string{
		bcgPage("0"):    bcgFixture(t, "listing.json"),
		bcgPage("500"):  bcgFixture(t, "listing-2.json"),
		bcgPage("1000"): `{"refineSearch":{"data":{"jobs":[]}}}`,
	}}
	got, err := (&portalSource{p: loadBCG(t), client: d}).Search(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	co, src, page := "Boston Consulting Group", "portal:bcg", "https://careers.bcg.com/global/en/job/"
	want := []Job{
		{ID: src + ":900201", Source: src, Title: "Consultant, Fictional Public Sector Practice", Company: co,
			Location: "Berlin, Berlin, Germany", URL: page + "900201"},
		{ID: src + ":900202", Source: src, Title: "Project Leader, Example Global Health", Company: co,
			Location: "Nairobi, Nairobi County, Kenya", URL: page + "900202"},
		{ID: src + ":900203", Source: src, Title: "Data Scientist, Imaginary Analytics Lab", Company: co,
			URL: page + "900203"},
		{ID: src + ":900205", Source: src, Title: "Associate, Invented Climate Practice", Company: co,
			Location: "Santiago, Santiago Metropolitan, Chile", URL: page + "900205"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
	if want := []string{bcgPage("0"), bcgPage("500"), bcgPage("1000")}; !reflect.DeepEqual(d.asked, want) {
		t.Errorf("asked %q\nwant  %q", d.asked, want)
	}
	r := d.reqs[0]
	if r.Method != http.MethodPost ||
		r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Accept") != "application/json" {
		t.Errorf("request: %s %v", r.Method, r.Header)
	}
}

// A BCG job page is recognised as the bcg portal, with its job ID and ats,
// with or without the title slug the site links to (ADR 0031).
func TestPortalBCGMatches(t *testing.T) {
	p := loadBCG(t)
	if got := p.Hosts(); !reflect.DeepEqual(got, []string{"careers.bcg.com"}) {
		t.Errorf("hosts = %v", got)
	}
	cases := map[string]string{
		"https://careers.bcg.com/global/en/job/900201":                  "portal:bcg:900201",
		"https://careers.bcg.com/global/en/job/900201#apply":            "portal:bcg:900201",
		"https://careers.bcg.com/global/en/job/900201/Consultant-Slug":  "portal:bcg:900201",
		"https://careers.bcg.com/global/en/job/900201/Consultant#apply": "portal:bcg:900201",
		"https://careers.bcg.com/global/en/job/900201/Consultant/apply": "",
		"https://careers.bcg.com/jp/ja/job/900201/Consultant-Slug":      "",
		"https://careers.bcg.com/global/en/search-results":              "",
		"https://experiencedtalent.bcg.com/global/en/job/900201":        "",
		"https://careers.bcg.com/jp/ja/job/900201":                      "",
	}
	for u, want := range cases {
		got, id, ok := PortalFor([]Portal{p}, u)
		if ok != (want != "") || id != want || ok && (got.Org != "bcg" || got.ATS != "phenom") {
			t.Errorf("%s: %q %v %q, want %q", u, got.Org, ok, id, want)
		}
	}
}
