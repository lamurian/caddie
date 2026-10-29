package main

import (
	"strings"
	"testing"
)

// whoHTTP serves the recorded WHO listing (path relative to testdata/sources/).
var whoHTTP = fakeHTTP{
	"https://careers.who.int/careersection/rest/jobboard/searchjobs?lang=en&portal=101430233": "../portals/who/listing.json",
}

// `caddie search --source portal:who` runs the tracked portals/who.yml
// against its recorded response; roles.include keeps the officer titles.
func TestSearchPortalWHO(t *testing.T) {
	prefs := "roles:\n  include: [officer]\nsources:\n  portals: [who]\n"
	// The path is relative to testdata/portals/, so this is the repo's portals/who.yml.
	vars := portalWorkspace(t, prefs, map[string]string{"who": "../../portals/who.yml"})
	r := runWith(t, vars, whoHTTP, "search", "--source", "portal:who")
	if r.code != exitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	want := `{
  "jobs": [
    {
      "id": "portal:who:2690101",
      "source": "portal:who",
      "title": "Technical Officer, Fictional Disease Surveillance",
      "company": "World Health Organization",
      "url": "https://careers.who.int/careersection/ex/jobdetail.ftl?job=2690101&lang=en",
      "location": "Switzerland-Geneva"
    },
    {
      "id": "portal:who:2690104",
      "source": "portal:who",
      "title": "Budget and Finance Officer (BFO)",
      "company": "World Health Organization",
      "url": "https://careers.who.int/careersection/ex/jobdetail.ftl?job=2690104&lang=en"
    }
  ],
  "hidden_seen": 0,
  "notices": []
}`
	if strings.TrimSpace(r.stdout) != want {
		t.Errorf("got\n%s\nwant\n%s", r.stdout, want)
	}
}
