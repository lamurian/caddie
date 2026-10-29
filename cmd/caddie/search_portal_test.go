package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/source"
)

// portalHTTP serves the recorded portal listings (paths are relative to
// testdata/sources/).
var portalHTTP = fakeHTTP{
	"https://careers.example.org/vacancies": "../portals/example-html/listing.html",
	"https://jobs.example.net/api/search":   "../portals/example-json/listing.json",
}

// portalWorkspace makes a workspace whose portals/ holds the given configs
// from testdata/portals/ (name → source file).
func portalWorkspace(t *testing.T, prefs string, configs map[string]string) map[string]string {
	t.Helper()
	vars := newWorkspace(t, prefs)
	dir := filepath.Join(vars["CADDIE_WORKSPACE"], "portals")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	for name, src := range configs {
		data, err := os.ReadFile(filepath.Join("../../testdata/portals", src))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name+".yml"), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return vars
}

const portalPrefs = `roles:
  include: [officer, economist, epidemiologist]
sources:
  remotive: {enabled: true}
  portals: [example-html, example-json, absent]
`

func TestSearchPortals(t *testing.T) {
	vars := portalWorkspace(t, portalPrefs,
		map[string]string{"example-html": "example-html.yml", "example-json": "example-json.yml"})
	r := runWith(t, vars, portalHTTP, "search", "--source", "portal:example-html")
	if r.code != exitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	want := `{
  "jobs": [
    {
      "id": "portal:example-html:EHO-2041",
      "source": "portal:example-html",
      "title": "Technical Officer, Health Data Systems",
      "company": "Example Health Organization",
      "url": "https://careers.example.org/job?id=EHO-2041",
      "location": "Geneva, Switzerland",
      "deadline": "2026-11-30"
    },
    {
      "id": "portal:example-html:EHO-2057",
      "source": "portal:example-html",
      "title": "Epidemiologist",
      "company": "Example Health Organization",
      "url": "https://careers.example.org/job?id=EHO-2057",
      "location": "Remote",
      "work_mode": "wfh",
      "deadline": "2026-12-15"
    }
  ],
  "hidden_seen": 0,
  "notices": []
}`
	if strings.TrimSpace(r.stdout) != want {
		t.Errorf("got\n%s\nwant\n%s", r.stdout, want)
	}
	res, r := searchWith(t, vars, "--source", "portal", "--include-seen")
	ids := jobIDs(res)
	wantIDs := []string{"portal:example-html:EHO-2041", "portal:example-html:EHO-2057",
		"portal:example-json:9007199254740993123", "portal:example-json:4502"}
	if r.code != exitOK || !reflect.DeepEqual(ids, wantIDs) {
		t.Errorf("--source portal: exit %d, %v, want %v", r.code, ids, wantIDs)
	}
	if n := strings.Join(res.Notices, "\n"); !strings.Contains(n, "portal:absent: skipped: ") ||
		!strings.Contains(n, "absent.yml does not exist") || strings.Contains(n, "remotive") {
		t.Errorf("notices: %v", res.Notices)
	}
	res, _ = searchWith(t, vars, "--source", "portal:who")
	if len(res.Jobs) != 0 || !strings.Contains(strings.Join(res.Notices, ""), "not in sources.portals") {
		t.Errorf("--source portal:who: %+v", res)
	}
}

func TestSearchInvalidPortalConfig(t *testing.T) {
	vars := portalWorkspace(t, "sources: {portals: [bad]}\n",
		map[string]string{"bad": "invalid/missing-listing-url.yml"})
	r := runWith(t, vars, portalHTTP, "search")
	if r.code != exitConfig || !strings.Contains(r.stderr, filepath.Join("portals", "bad.yml")+": listing_url: required") {
		t.Errorf("got %+v", r)
	}
	// Not selected by --source, the broken config is not loaded.
	if r = runWith(t, vars, portalHTTP, "search", "--source", "remotive"); r.code == exitConfig {
		t.Errorf("--source remotive loaded portals: %+v", r)
	}
}

func searchWith(t *testing.T, vars map[string]string, args ...string) (source.Result, result) {
	t.Helper()
	r := runWith(t, vars, portalHTTP, append([]string{"search"}, args...)...)
	var res source.Result
	if r.code == exitOK {
		if err := json.Unmarshal([]byte(r.stdout), &res); err != nil {
			t.Fatalf("%v\n%s", err, r.stdout)
		}
	}
	return res, r
}
