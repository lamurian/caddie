package source

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

const (
	bainConfig = "../../portals/bain.yml"
	bainJob    = "https://www.bain.com/careers/find-a-role/position/?jobid=70101"
)

func loadBain(t *testing.T) Portal {
	t.Helper()
	p, err := LoadPortal(bainConfig)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// The tracked portals/bain.yml is discovery-only: its listing API refuses
// direct access, so search skips it with a notice and sends no request.
func TestPortalBainSkipped(t *testing.T) {
	p := loadBain(t)
	if p.SkipReason == "" || p.Name != "Bain & Company" || p.ATS != "bespoke" {
		t.Fatalf("config: %+v", p)
	}
	prefs := allEnabled()
	prefs.Portals = []string{"bain"}
	d := &portalDoer{t: t}
	srcs, notices := Build(prefs, Keys{"id", "key"}, "portal:bain", d, []Portal{p})
	want := "portal:bain: skipped: the listing API refuses direct access (see portals/bain.yml notes)"
	if len(srcs) != 0 || len(notices) != 1 || notices[0] != want || len(d.reqs) != 0 {
		t.Errorf("sources %v, notices %v, requests %d", names(srcs), notices, len(d.reqs))
	}
}

// A Bain job page is recognised as the bain portal, with its job ID and ats.
func TestPortalBainMatches(t *testing.T) {
	p := loadBain(t)
	if got := p.Hosts(); !reflect.DeepEqual(got, []string{"www.bain.com"}) {
		t.Errorf("hosts = %v", got)
	}
	cases := map[string]string{
		bainJob:            "portal:bain:70101",
		bainJob + "#apply": "portal:bain:70101",
		"https://www.bain.com/careers/find-a-role/position/?jobid=":         "",
		"https://www.bain.com/careers/find-a-role/":                         "",
		"https://www.bain.com/de/careers/find-a-role/position/?jobid=70101": "",
		"https://bain.com/careers/find-a-role/position/?jobid=70101":        "",
	}
	for u, want := range cases {
		got, id, ok := PortalFor([]Portal{p}, u)
		if ok != (want != "") || id != want || ok && (got.Org != "bain" || got.ATS != "bespoke") {
			t.Errorf("%s: %q %v %q, want %q", u, got.Org, ok, id, want)
		}
	}
}

// A saved Bain job page yields its posting from the JSON-LD block, although
// the visible page is an empty React mount.
func TestPortalBainDetail(t *testing.T) {
	d := &portalDoer{t: t, routes: map[string]string{bainJob: "bain/detail-70101.html"}}
	got, err := Fetch(context.Background(), d, bainJob)
	if err != nil {
		t.Fatal(err)
	}
	j := got.Job
	if j.Title != "Associate Consultant, Fictional Practice" || j.Company != "Bain & Company" ||
		j.Location != "Fictionville, Imaginaria" || !strings.Contains(got.HTML, "imaginary strategy") {
		t.Errorf("detail: %+v", got)
	}
}
