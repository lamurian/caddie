package source

import (
	"reflect"
	"testing"
)

const (
	mckinseyConfig = "../../portals/mckinsey.yml"
	mckinseyJobs   = "https://www.mckinsey.com/careers/search-jobs/jobs/"
)

func loadMcKinsey(t *testing.T) Portal {
	t.Helper()
	p, err := LoadPortal(mckinseyConfig)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// The tracked portals/mckinsey.yml is discovery-only: McKinsey's terms bar
// automated collection, so search skips it with a notice and sends no request.
func TestPortalMcKinseySkipped(t *testing.T) {
	p := loadMcKinsey(t)
	if p.SkipReason == "" || p.Name != "McKinsey & Company" || p.ATS != "avature" ||
		p.JSONPath != "" || p.ItemSelector != "" {
		t.Fatalf("config: %+v", p)
	}
	prefs := allEnabled()
	prefs.Portals = []string{"mckinsey"}
	d := &portalDoer{t: t}
	srcs, notices := Build(prefs, Keys{"id", "key"}, "portal:mckinsey", d, []Portal{p})
	want := "portal:mckinsey: skipped: the terms forbid automated collection (see portals/mckinsey.yml notes)"
	if len(srcs) != 0 || len(notices) != 1 || notices[0] != want || len(d.reqs) != 0 {
		t.Errorf("sources %v, notices %v, requests %d", names(srcs), notices, len(d.reqs))
	}
}

// A McKinsey job page is recognised as the mckinsey portal, with its
// friendly-URL slug as the job ID and avature as the ats.
func TestPortalMcKinseyMatches(t *testing.T) {
	p := loadMcKinsey(t)
	if got := p.Hosts(); !reflect.DeepEqual(got, []string{"www.mckinsey.com"}) {
		t.Errorf("hosts = %v", got)
	}
	cases := map[string]string{
		mckinseyJobs + "fictionalanalyst-90101":       "portal:mckinsey:fictionalanalyst-90101",
		mckinseyJobs + "fictionalanalyst-90101#apply": "portal:mckinsey:fictionalanalyst-90101",
		mckinseyJobs: "",
		"https://www.mckinsey.com/careers/search-jobs":                           "",
		"https://mckinsey.com/careers/search-jobs/jobs/x":                        "",
		"https://mckinsey.avature.net/careers/ApplicationMethods?folderId=90101": "",
	}
	for u, want := range cases {
		got, id, ok := PortalFor([]Portal{p}, u)
		if ok != (want != "") || id != want || ok && (got.Org != "mckinsey" || got.ATS != "avature") {
			t.Errorf("%s: %q %v %q, want %q", u, got.Org, ok, id, want)
		}
	}
}
