package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/memory"
	"github.com/lamurian/caddie/internal/source"
)

const searchPrefs = `roles:
  include: [data scientist, biostatistician]
  exclude: [sales engineer]
companies:
  exclude: [Vandelay]
sources:
  adzuna: {enabled: true, countries: [nl]}
  remotive: {enabled: true}
  arbeitnow: {enabled: false}
  greenhouse_boards: [acme]
  lever_boards: [globex]
`

var boards = fakeHTTP{
	"https://remotive.com/api/remote-jobs?search=data+scientist":                     "remotive.json",
	"https://remotive.com/api/remote-jobs?search=biostatistician":                    "remotive.json",
	"https://boards-api.greenhouse.io/v1/boards/acme/jobs?content=true":              "greenhouse-acme.json",
	"https://api.lever.co/v0/postings/globex?mode=json":                              "lever-globex.json",
	"https://boards-api.greenhouse.io/v1/boards/acme/jobs/123?pay_transparency=true": "greenhouse-acme-123.json",
}

// newWorkspace makes an empty workspace with preferences.yml and returns
// the env vars pointing at it.
func newWorkspace(t *testing.T, prefs string) map[string]string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "preferences.yml"), []byte(prefs), 0o600); err != nil {
		t.Fatal(err)
	}
	return map[string]string{"CADDIE_WORKSPACE": root}
}

func search(t *testing.T, vars map[string]string, args ...string) (source.Result, result) {
	t.Helper()
	r := runWith(t, vars, boards, append([]string{"search"}, args...)...)
	var res source.Result
	if r.code == exitOK {
		if err := json.Unmarshal([]byte(r.stdout), &res); err != nil {
			t.Fatalf("%v\n%s", err, r.stdout)
		}
	}
	return res, r
}

func jobIDs(res source.Result) []string {
	out := []string{}
	for _, j := range res.Jobs {
		out = append(out, j.ID)
	}
	return out
}

// Disabled sources and missing Adzuna keys are notices, not errors; the
// excluded company is dropped and duplicate URLs appear once.
func TestSearchCommand(t *testing.T) {
	vars := newWorkspace(t, searchPrefs)
	res, r := search(t, vars)
	if r.code != exitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	want := []string{"remotive:7000001", "greenhouse-board:123", "lever-board:0f8e7d6c-5b4a-4321-9876-0123456789ab"}
	if got := jobIDs(res); !reflect.DeepEqual(got, want) {
		t.Errorf("jobs = %v, want %v", got, want)
	}
	notices := strings.Join(res.Notices, "\n")
	for _, n := range []string{"adzuna: skipped: CADDIE_ADZUNA_APP_ID", "arbeitnow: skipped: disabled"} {
		if !strings.Contains(notices, n) {
			t.Errorf("notices lack %q: %v", n, res.Notices)
		}
	}
	if _, err := os.Stat(filepath.Join(vars["CADDIE_WORKSPACE"], ".caddie", memory.LastSearchFile)); err != nil {
		t.Errorf("last search not remembered: %v", err)
	}
	res, _ = search(t, vars, "--limit", "1", "--source", "lever-board")
	if got := jobIDs(res); len(got) != 1 || !strings.HasPrefix(got[0], "lever-board:") {
		t.Errorf("--source lever-board --limit 1: %v", got)
	}
	res, r = search(t, vars, "--source", "arbeitnow")
	if r.code != exitOK || len(res.Jobs) != 0 || !strings.Contains(strings.Join(res.Notices, ""), "arbeitnow") {
		t.Errorf("disabled --source: %+v %+v", r, res)
	}
}

func TestSearchExcludesSeenUnlessIncluded(t *testing.T) {
	vars := newWorkspace(t, searchPrefs)
	search(t, vars)
	r := runWith(t, vars, boards, "seen", "add", "remotive:7000001", "greenhouse-board:123", "--shown")
	if r.code != exitOK || !strings.Contains(r.stdout, `"recorded": [`) {
		t.Fatalf("seen add: %+v", r)
	}
	res, _ := search(t, vars)
	if got := jobIDs(res); len(got) != 1 || res.HiddenSeen != 2 {
		t.Errorf("seen jobs not excluded: %v, hidden %d", got, res.HiddenSeen)
	}
	res, _ = search(t, vars, "--include-seen")
	if got := jobIDs(res); len(got) != 3 || res.HiddenSeen != 0 {
		t.Errorf("--include-seen: %v, hidden %d", got, res.HiddenSeen)
	}
}

func TestSearchFailsWhenEverySourceFails(t *testing.T) {
	vars := newWorkspace(t, "sources: {remotive: {enabled: true}}\n")
	r := runCaddie(t, vars, "search")
	if r.code != exitExternal || !strings.Contains(r.stderr, "caddie search: fetch failed") {
		t.Errorf("got %+v", r)
	}
}

func TestSeenListReportsHandEdits(t *testing.T) {
	vars := newWorkspace(t, "")
	state := filepath.Join(vars["CADDIE_WORKSPACE"], ".caddie")
	lines := `{"id":"a:1","status":"shown"}` + "\n\n" + `oops` + "\n" + `{"id":"a:2","status":"rejected"}` + "\n"
	if err := os.MkdirAll(state, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, "seen.jsonl"), []byte(lines), 0o600); err != nil {
		t.Fatal(err)
	}
	var out seenListOutput
	r := runCaddie(t, vars, "seen", "list")
	if err := json.Unmarshal([]byte(r.stdout), &out); err != nil || r.code != exitOK {
		t.Fatalf("exit %d, %v: %s", r.code, err, r.stdout)
	}
	if len(out.Entries) != 2 || len(out.Problems) != 2 || out.Problems[0].Line != 2 || out.Problems[1].Line != 3 {
		t.Errorf("got %+v", out)
	}
	r = runCaddie(t, vars, "seen", "list", "--rejected")
	if !strings.Contains(r.stdout, `"a:2"`) || strings.Contains(r.stdout, `"a:1"`) {
		t.Errorf("--rejected: %s", r.stdout)
	}
}
