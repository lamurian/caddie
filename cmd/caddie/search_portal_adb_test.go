package main

import (
	"strings"
	"testing"
)

// adbHTTP serves the recorded ADB listing (path relative to testdata/sources/).
var adbHTTP = fakeHTTP{
	"https://www.adb.org/work-with-us/careers/current-opportunities": "../portals/adb/listing.html",
}

// `caddie search --source portal:adb` runs the tracked portals/adb.yml
// against its recorded page; roles.include keeps the officer title.
func TestSearchPortalADB(t *testing.T) {
	prefs := "roles:\n  include: [officer]\nsources:\n  portals: [adb]\n"
	// The path is relative to testdata/portals/, so this is the repo's portals/adb.yml.
	vars := portalWorkspace(t, prefs, map[string]string{"adb": "../../portals/adb.yml"})
	r := runWith(t, vars, adbHTTP, "search", "--source", "portal:adb")
	if r.code != exitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	want := `{
  "jobs": [
    {
      "id": "portal:adb:990104",
      "source": "portal:adb",
      "title": "Finance and Administration Officer / 990104",
      "company": "Asian Development Bank",
      "url": "https://www.adb.org/careers/990104",
      "deadline": "2026-11-04"
    }
  ],
  "hidden_seen": 0,
  "notices": []
}`
	if strings.TrimSpace(r.stdout) != want {
		t.Errorf("got\n%s\nwant\n%s", r.stdout, want)
	}
}
