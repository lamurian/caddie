package posting

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/memory"
	"github.com/lamurian/caddie/internal/source"
	"github.com/lamurian/caddie/internal/workspace"
)

var update = flag.Bool("update", false, "rewrite the golden posting in testdata/golden")

const (
	golden   = "../../testdata/golden/posting-senior-data-scientist-at-acme.md"
	acmeURL  = "https://boards.greenhouse.io/acme/jobs/123"
	acmeAPI  = "https://boards-api.greenhouse.io/v1/boards/acme/jobs/123?pay_transparency=true"
	plainURL = "https://www.example.net/rs"
)

// fakeDoer serves recorded responses from testdata/sources/, or inline
// HTML when the value starts with "<"; other URLs 404.
type fakeDoer map[string]string

func (f fakeDoer) Do(req *http.Request) (*http.Response, error) {
	file, ok := f[req.URL.String()]
	if !ok {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewReader(nil))}, nil
	}
	if strings.HasPrefix(file, "<") {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(file))}, nil
	}
	data, err := os.ReadFile(filepath.Join("../../testdata/sources", file))
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(data))}, err
}

var routes = fakeDoer{acmeAPI: "greenhouse-acme-123.json", plainURL: "page-plain.html"}

// newOptions builds an empty workspace, with the Acme review when withReview.
func newOptions(t *testing.T, withReview bool) Options {
	t.Helper()
	root := t.TempDir()
	l := workspace.New(config.Config{Workspace: root, StateDir: ".caddie"})
	if withReview {
		data, err := os.ReadFile("../../testdata/reviews/acme.md")
		if err != nil {
			t.Fatal(err)
		}
		mustWrite(t, filepath.Join(l.Dir(workspace.DirReviews), "acme.md"), string(data))
	}
	return Options{Layout: l, Store: memory.Store{Dir: l.State}, Client: routes, Today: "2026-10-28"}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSaveGolden(t *testing.T) {
	o := newOptions(t, true)
	res, err := Save(context.Background(), acmeURL, o)
	if err != nil {
		t.Fatal(err)
	}
	if res.Slug != "senior-data-scientist-at-acme" || !res.Created || !res.Review || res.ID != "greenhouse-board:123" {
		t.Errorf("result = %+v", res)
	}
	got, err := os.ReadFile(res.Path)
	if err != nil {
		t.Fatal(err)
	}
	if *update {
		mustWrite(t, golden, string(got))
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("posting differs from %s; rerun with -update and review the diff\n%s", golden, got)
	}
	checkSaved(t, o, res.Path, got)
}

// checkSaved asserts the frontmatter parses as status saved and that the
// job is marked saved in seen.jsonl.
func checkSaved(t *testing.T, o Options, path string, data []byte) {
	t.Helper()
	p, _, err := workspace.ParsePosting(path, data)
	if err != nil || p.Status != workspace.StatusSaved || p.ATS != "greenhouse" || p.SavedAt != "2026-10-28" {
		t.Errorf("frontmatter %+v, %v", p, err)
	}
	log, _ := o.Store.Log()
	if e := log.Latest(); len(e) != 1 || e[0].ID != "greenhouse-board:123" || e[0].Status != memory.Saved {
		t.Errorf("seen.jsonl = %+v", e)
	}
}

func TestSaveWithoutReview(t *testing.T) {
	o := newOptions(t, false)
	res, err := Save(context.Background(), acmeURL, o)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(res.Path)
	if res.Review || !strings.Contains(string(data), "## Reviews\n\n_No reviews/acme.md at save time._\n") {
		t.Errorf("review section:\n%s", data)
	}
}

// Saving the same URL twice is stable; a different posting with the same
// title and company gets the next free slug.
func TestSlugsAreStableAndUnique(t *testing.T) {
	o := newOptions(t, false)
	mustWrite(t, filepath.Join(o.Layout.Dir(workspace.DirApplied), "senior-data-scientist-at-acme.md"),
		"---\nslug: senior-data-scientist-at-acme\nurl: https://boards.greenhouse.io/acme/jobs/1\nstatus: applied\n---\n")
	first, err := Save(context.Background(), acmeURL, o)
	if err != nil || first.Slug != "senior-data-scientist-at-acme-2" {
		t.Fatalf("first save: %+v, %v", first, err)
	}
	before, _ := os.ReadFile(first.Path)
	again, err := Save(context.Background(), acmeURL, o)
	if err != nil || again.Slug != first.Slug || again.Created || again.ID != first.ID {
		t.Errorf("second save: %+v, %v", again, err)
	}
	if after, _ := os.ReadFile(first.Path); !bytes.Equal(before, after) {
		t.Error("saving again must not rewrite the posting")
	}
}

// A job remembered from the last search keeps its ID, source, title,
// company and salary even when the page itself states less.
func TestSaveUsesRememberedJob(t *testing.T) {
	o := newOptions(t, false)
	known := source.Job{ID: "adzuna:42", Source: source.Adzuna, Title: "Research Scientist (Remote)",
		Company: "Cyberdyne Systems", URL: plainURL, SalaryMin: 50000, SalaryMax: 60000,
		SalaryCurrency: "EUR", SalaryPeriod: "year"}
	if err := o.Store.Remember([]source.Job{known}); err != nil {
		t.Fatal(err)
	}
	res, err := Save(context.Background(), plainURL, o)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(res.Path)
	p, body, err := workspace.ParsePosting(res.Path, data)
	if err != nil || res.ID != "adzuna:42" || res.Slug != "research-scientist-remote-at-cyberdyne-systems" {
		t.Fatalf("result %+v, %v", res, err)
	}
	if p.Source != "adzuna" || p.SalaryMin != 50000 || p.Deadline != "2026-12-05" || p.ATS != "" ||
		p.ATSHost != "www.example.net" {
		t.Errorf("frontmatter = %+v", p)
	}
	for _, want := range []string{"## Description\n\nCyberdyne is", "## Responsibilities\n\n**Your responsibilities**\n\n- Run experiments.", "## Requirements\n\n- PhD",
		"## Perks and benefits\n\n**Benefits**\n\n- Pension plan."} {
		if !strings.Contains(string(body), want) {
			t.Errorf("body lacks %q:\n%s", want, body)
		}
	}
}

func TestSaveErrors(t *testing.T) {
	o := newOptions(t, false)
	o.Client = fakeDoer{"https://example.org/empty": "<html><body><p>Cookies?</p></body></html>"}
	if _, err := Save(context.Background(), "https://example.org/gone", o); !errors.Is(err, source.ErrFetch) {
		t.Errorf("404: got %v", err)
	}
	if _, err := Save(context.Background(), "https://example.org/empty", o); !errors.Is(err, ErrNoPosting) {
		t.Errorf("no title: got %v", err)
	}
}
