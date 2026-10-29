package main

import (
	"strings"
	"testing"
)

// unHTTP serves the recorded UN listing (path relative to testdata/sources/).
var unHTTP = fakeHTTP{
	"https://careers.un.org/api/public/opening/jo/list/filteredV2/en": "../portals/un/listing.json",
}

// `caddie search --source portal:un` runs the tracked portals/un.yml
// against its recorded response; roles.include keeps the officer titles.
func TestSearchPortalUN(t *testing.T) {
	prefs := "roles:\n  include: [officer]\nsources:\n  portals: [un]\n"
	// The path is relative to testdata/portals/, so this is the repo's portals/un.yml.
	vars := portalWorkspace(t, prefs, map[string]string{"un": "../../portals/un.yml"})
	r := runWith(t, vars, unHTTP, "search", "--source", "portal:un")
	if r.code != exitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	want := `{
  "jobs": [
    {
      "id": "portal:un:900101",
      "source": "portal:un",
      "title": "Programme Management Officer, Fictional Affairs",
      "company": "United Nations",
      "url": "https://careers.un.org/jobSearchDescription/900101?language=en",
      "deadline": "2026-11-28"
    },
    {
      "id": "portal:un:900103",
      "source": "portal:un",
      "title": "Associate Economic Affairs Officer",
      "company": "United Nations",
      "url": "https://careers.un.org/jobSearchDescription/900103?language=en"
    }
  ],
  "hidden_seen": 0,
  "notices": []
}`
	if strings.TrimSpace(r.stdout) != want {
		t.Errorf("got\n%s\nwant\n%s", r.stdout, want)
	}
}
