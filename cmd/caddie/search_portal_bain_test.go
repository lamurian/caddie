package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const bainJob = "https://www.bain.com/careers/find-a-role/position/?jobid=70101"

// bainHTTP serves only the recorded Bain job page (path relative to
// testdata/sources/): search must not call the listing at all.
var bainHTTP = fakeHTTP{bainJob: "../portals/bain/detail-70101.html"}

// The tracked portals/bain.yml is discovery-only: `caddie search` skips it
// with a notice, and `caddie posting save` on a job page takes the portal's
// source, ID and ats.
func TestSearchPortalBain(t *testing.T) {
	prefs := "sources:\n  portals: [bain]\n"
	vars := portalWorkspace(t, prefs, map[string]string{"bain": "../../portals/bain.yml"})
	r := runWith(t, vars, bainHTTP, "search", "--source", "portal:bain")
	if r.code != exitOK || !strings.Contains(r.stdout, `"jobs": []`) ||
		!strings.Contains(r.stdout, "portal:bain: skipped: the listing API refuses direct access") {
		t.Fatalf("search: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	r = runWith(t, vars, bainHTTP, "posting", "save", bainJob)
	if r.code != exitOK || !strings.Contains(r.stdout, `"id": "portal:bain:70101"`) {
		t.Fatalf("save: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	got, _ := os.ReadFile(filepath.Join(vars["CADDIE_WORKSPACE"], "postings",
		"associate-consultant-fictional-practice-at-bain-company.md"))
	for _, want := range []string{"source: portal:bain\n", "ats: bespoke\n", "ats_host: www.bain.com\n"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("posting lacks %q:\n%s", want, got)
		}
	}
}
