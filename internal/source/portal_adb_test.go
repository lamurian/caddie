package source

import (
	"net/http"
	"reflect"
	"testing"
)

const (
	adbConfig  = "../../portals/adb.yml"
	adbListing = "https://www.adb.org/work-with-us/careers/current-opportunities"
)

func loadADB(t *testing.T) Portal {
	t.Helper()
	p, err := LoadPortal(adbConfig)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// The tracked portals/adb.yml parses the saved ADB listing across its
// per-group tables: header rows are skipped, the untitled row is dropped, an
// empty location stays empty, and the closing date reads with or without a
// <time> element. The ID is the reference number detail_url captures; the
// title keeps its " / <number>" suffix.
func TestPortalADB(t *testing.T) {
	d := &portalDoer{t: t, routes: map[string]string{adbListing: "adb/listing.html"}}
	got := searchPortal(t, loadADB(t), d)
	co, src, page := "Asian Development Bank", "portal:adb", "https://www.adb.org/careers/"
	want := []Job{
		{ID: src + ":990101", Source: src, Title: "Senior Fictional Infrastructure Specialist / 990101",
			Company: co, URL: page + "990101", Location: "Asian Development Bank Headquarters",
			Deadline: "2026-11-03"},
		{ID: src + ":990102", Source: src, Title: "Economist (Imaginary Research) / 990102",
			Company: co, URL: page + "990102", Location: "Example Resident Mission", Deadline: "2026-12-28"},
		{ID: src + ":990104", Source: src, Title: "Finance and Administration Officer / 990104",
			Company: co, URL: page + "990104", Deadline: "2026-11-04"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
	r := d.reqs[0]
	if len(d.reqs) != 1 || r.Method != http.MethodGet || d.bodies[0] != "" ||
		r.Header.Get("Accept") != "text/html" {
		t.Errorf("request: %s %v %q", r.Method, r.Header, d.bodies[0])
	}
}

// An ADB job page is recognised as the adb portal, with its job ID and ats;
// the ACES apply page is not a job page.
func TestPortalADBMatches(t *testing.T) {
	p := loadADB(t)
	if got := p.Hosts(); !reflect.DeepEqual(got, []string{"www.adb.org"}) {
		t.Errorf("hosts = %v", got)
	}
	cases := map[string]string{
		"https://www.adb.org/careers/990101":            "portal:adb:990101",
		"https://www.adb.org/careers/990101#apply":      "portal:adb:990101",
		"https://www.adb.org/careers/990101/extra":      "",
		"https://www.adb.org/work-with-us/careers":      "",
		"https://aces.adb.org/?job=990101":              "",
		"https://www.adb.org/work-with-us/careers/faqs": "",
	}
	for u, want := range cases {
		got, id, ok := PortalFor([]Portal{p}, u)
		if ok != (want != "") || id != want || ok && (got.Org != "adb" || got.ATS != "bespoke") {
			t.Errorf("%s: %q %v %q, want %q", u, got.Org, ok, id, want)
		}
	}
}
