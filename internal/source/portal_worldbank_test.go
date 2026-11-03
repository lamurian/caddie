package source

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const worldbankConfig = "../../portals/worldbank.yml"

func loadWorldBank(t *testing.T) Portal {
	t.Helper()
	p, err := LoadPortal(worldbankConfig)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// The saved Cornerstone career site is an empty JavaScript shell: its list
// comes from a token-authorised API that ADR 0019 cannot call. So the
// tracked portals/worldbank.yml is js_rendered: search skips it with a
// notice and sends no request.
func TestPortalWorldBank(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(portalDir, "worldbank/listing.html"))
	if err != nil {
		t.Fatal(err)
	}
	if page := string(data); !strings.Contains(page, `<div id="cs-root"></div>`) ||
		!strings.Contains(page, `"token":`) || strings.Contains(page, "requisition") {
		t.Errorf("fixture is not an empty career-site shell")
	}
	p := loadWorldBank(t)
	if !p.JSRendered || p.Name != "World Bank Group" || p.ATS != "cornerstone" {
		t.Errorf("config: %+v", p)
	}
	d := &portalDoer{t: t, routes: map[string]string{p.ListingURL: "worldbank/listing.html"}}
	srcs, notices := portalSources([]string{"worldbank"}, []Portal{p}, "portal:worldbank", d)
	want := "portal:worldbank: skipped: listing needs JavaScript (see portals/worldbank.yml notes)"
	if len(srcs) != 0 || !reflect.DeepEqual(notices, []string{want}) || len(d.reqs) != 0 {
		t.Errorf("sources %v, notices %v, requests %d", names(srcs), notices, len(d.reqs))
	}
}

// A World Bank Group job page is recognised as the worldbank portal, with
// its requisition ID and ats, though search skips the portal.
func TestPortalWorldBankMatches(t *testing.T) {
	p := loadWorldBank(t)
	if got := p.Hosts(); !reflect.DeepEqual(got, []string{"worldbankgroup.csod.com"}) {
		t.Errorf("hosts = %v", got)
	}
	page := "https://worldbankgroup.csod.com/ux/ats/careersite/1/home/requisition/"
	cases := map[string]string{
		page + "40123?c=worldbankgroup":                                                          "portal:worldbank:40123",
		page + "40123?c=worldbankgroup#details":                                                  "portal:worldbank:40123",
		page + "40123?c=othercorp":                                                               "",
		"https://worldbankgroup.csod.com/ux/ats/careersite/1/home?c=worldbankgroup":              "",
		"https://othercorp.csod.com/ux/ats/careersite/1/home/requisition/40123?c=worldbankgroup": "",
	}
	for u, want := range cases {
		got, id, ok := PortalFor([]Portal{p}, u)
		if ok != (want != "") || id != want || ok && (got.Org != "worldbank" || got.ATS != "cornerstone") {
			t.Errorf("%s: %q %v %q, want %q", u, got.Org, ok, id, want)
		}
	}
}
