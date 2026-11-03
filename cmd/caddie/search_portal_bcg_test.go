package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// bcgSlugJob is a BCG job page as the site links it, with a title slug.
const bcgSlugJob = "https://careers.bcg.com/global/en/job/900201/Consultant-Fictional-Public-Sector-Practice"

// `caddie posting save` on a slugged BCG job page takes the bcg portal's
// source, ats and the same ID search gives that job (ADR 0031).
func TestPostingSavePortalBCGSlug(t *testing.T) {
	vars := portalWorkspace(t, "", map[string]string{"bcg": "../../portals/bcg.yml"})
	bcgHTTP := fakeHTTP{bcgSlugJob: "../portals/bcg/detail-900201.html"}
	r := runWith(t, vars, bcgHTTP, "posting", "save", bcgSlugJob)
	if r.code != exitOK || !strings.Contains(r.stdout, `"id": "portal:bcg:900201"`) {
		t.Fatalf("save: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	got, _ := os.ReadFile(filepath.Join(vars["CADDIE_WORKSPACE"], "postings",
		"consultant-fictional-public-sector-practice-at-boston-consulting-group.md"))
	for _, want := range []string{"source: portal:bcg\n", "ats: phenom\n", "ats_host: careers.bcg.com\n"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("posting lacks %q:\n%s", want, got)
		}
	}
}
