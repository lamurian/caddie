package posting

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/memory"
	"github.com/lamurian/caddie/internal/source"
	"github.com/lamurian/caddie/internal/workspace"
)

const portalURL = "https://careers.example.org/job?id=1001"

// portalOptions is an empty workspace that knows the example-html portal
// and serves its saved detail page.
func portalOptions(t *testing.T) Options {
	t.Helper()
	p, err := source.LoadPortal("../../testdata/portals/example-html.yml")
	if err != nil {
		t.Fatal(err)
	}
	page, err := os.ReadFile("../../testdata/portals/example-html/detail-1001.html")
	if err != nil {
		t.Fatal(err)
	}
	o := newOptions(t, false)
	o.Client = fakeDoer{portalURL: string(page), plainURL: "page-plain.html"}
	o.Portals = []source.Portal{p}
	return o
}

func savedPosting(t *testing.T, path string) (workspace.Posting, string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p, body, err := workspace.ParsePosting(path, data)
	if err != nil {
		t.Fatal(err)
	}
	return p, string(body)
}

// A portal URL takes the portal's source, ID, name and ats (ADR 0019).
func TestSavePortalURL(t *testing.T) {
	o := portalOptions(t)
	res, err := Save(context.Background(), portalURL, o)
	if err != nil {
		t.Fatal(err)
	}
	p, body := savedPosting(t, res.Path)
	if res.ID != "portal:example-html:1001" ||
		res.Slug != "technical-officer-health-data-analytics-at-example-health-organization" {
		t.Errorf("result = %+v", res)
	}
	if p.Source != "portal:example-html" || p.Company != "Example Health Organization" ||
		p.ATS != "taleo" || p.ATSHost != "careers.example.org" || p.Deadline != "2026-12-15" {
		t.Errorf("frontmatter = %+v", p)
	}
	for _, want := range []string{"## Responsibilities\n\n**Duties and Responsibilities**",
		"## Requirements\n\n**Qualifications**", "## Perks and benefits\n\n**What we offer**"} {
		if !strings.Contains(body, want) {
			t.Errorf("body lacks %q:\n%s", want, body)
		}
	}
	checkSavedAgain(t, o, res)
}

// checkSavedAgain asserts seen.jsonl holds the saved ID and that saving
// the URL again keeps that ID without rewriting the posting.
func checkSavedAgain(t *testing.T, o Options, res Result) {
	t.Helper()
	log, _ := o.Store.Log()
	if e := log.Latest(); len(e) != 1 || e[0].ID != res.ID || e[0].Status != memory.Saved {
		t.Errorf("seen.jsonl = %+v", e)
	}
	again, err := Save(context.Background(), portalURL, o)
	if err != nil || again.Created || again.ID != res.ID {
		t.Errorf("second save: %+v, %v", again, err)
	}
}

// A URL no portal matches is saved as before: a web page, no ats.
func TestSaveNonPortalURL(t *testing.T) {
	o := portalOptions(t)
	res, err := Save(context.Background(), plainURL, o)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := savedPosting(t, res.Path)
	if !strings.HasPrefix(res.ID, "url:") || p.Source != source.Web || p.ATS != "" ||
		p.ATSHost != "www.example.net" {
		t.Errorf("result %+v, frontmatter %+v", res, p)
	}
}

// seen.jsonl alone (no last search) restores a portal job's source.
func TestSaveRecallsPortalSourceFromSeen(t *testing.T) {
	o := newOptions(t, false)
	err := o.Store.Record(memory.Entry{ID: "portal:who:77", URL: plainURL, Title: "Epidemiologist",
		Company: "World Health Organization", Status: memory.Shown, At: "2026-10-27"})
	if err != nil {
		t.Fatal(err)
	}
	res, err := Save(context.Background(), plainURL, o)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := savedPosting(t, res.Path)
	if res.ID != "portal:who:77" || p.Source != "portal:who" || p.Company != "World Health Organization" {
		t.Errorf("result %+v, frontmatter %+v", res, p)
	}
}
