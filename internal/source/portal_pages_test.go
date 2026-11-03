package source

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// pageDoer answers a paginated portal: pages maps the request's URL+body to
// an inline JSON response; anything else fails. It records each key asked.
type pageDoer struct {
	pages map[string]string
	asked []string
}

func (d *pageDoer) Do(req *http.Request) (*http.Response, error) {
	key := req.URL.String()
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		key += " " + string(b)
	}
	d.asked = append(d.asked, key)
	resp, ok := d.pages[key]
	if !ok {
		return nil, errors.New("connection refused")
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(resp))}, nil
}

// jobsJSON is a listing response holding one item per id.
func jobsJSON(ids ...string) string {
	var items []string
	for _, id := range ids {
		items = append(items, `{"id":"`+id+`","title":"Officer `+id+`"}`)
	}
	return `{"jobs":[` + strings.Join(items, ",") + `]}`
}

func pagedPortal(t *testing.T, extra string) Portal {
	t.Helper()
	path := filepath.Join(t.TempDir(), "paged.yml")
	data := "name: Paged\nats: bespoke\njson_path: jobs\nfields: {id: id, title: title}\n" +
		"detail_url: https://jobs.example.org/job/{id}\n" + extra
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	p, err := LoadPortal(path)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func pagedIDs(t *testing.T, p Portal, d *pageDoer) []string {
	t.Helper()
	jobs, err := (&portalSource{p: p, client: d}).Search(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	for _, j := range jobs {
		ids = append(ids, strings.TrimPrefix(j.ID, "portal:paged:"))
	}
	return ids
}

// A POST body's {page} counts from 1; a page that repeats an earlier one
// (Taleo past the end) stops paging, and items seen before are dropped.
func TestPortalPagesBodyStopsOnRepeat(t *testing.T) {
	u := "https://jobs.example.org/api"
	p := pagedPortal(t, "listing_url: "+u+"\nmethod: post\nbody: '{\"pageNo\":{page}}'\npages: 5\n")
	d := &pageDoer{pages: map[string]string{
		u + ` {"pageNo":1}`: jobsJSON("1", "2"),
		u + ` {"pageNo":2}`: jobsJSON("3", "2"),
		u + ` {"pageNo":3}`: jobsJSON("3"),
	}}
	if got := pagedIDs(t, p, d); !reflect.DeepEqual(got, []string{"1", "2", "3"}) {
		t.Errorf("ids = %v", got)
	}
	if len(d.asked) != 3 {
		t.Errorf("asked %q, want pages 1-3", d.asked)
	}
}

// A GET listing_url's {page} is filled the same way; an empty page stops
// paging, and `pages` caps the requests.
func TestPortalPagesQuery(t *testing.T) {
	u := "https://jobs.example.org/api?page="
	empty := &pageDoer{pages: map[string]string{u + "1": jobsJSON("1"), u + "2": jobsJSON()}}
	p := pagedPortal(t, "listing_url: "+u+"{page}\npages: 4\n")
	if got := pagedIDs(t, p, empty); !reflect.DeepEqual(got, []string{"1"}) || len(empty.asked) != 2 {
		t.Errorf("empty page: ids %v, asked %q", got, empty.asked)
	}
	capped := &pageDoer{pages: map[string]string{u + "1": jobsJSON("1"), u + "2": jobsJSON("2"), u + "3": jobsJSON("3")}}
	p = pagedPortal(t, "listing_url: "+u+"{page}\npages: 2\n")
	if got := pagedIDs(t, p, capped); !reflect.DeepEqual(got, []string{"1", "2"}) || len(capped.asked) != 2 {
		t.Errorf("capped: ids %v, asked %q", got, capped.asked)
	}
}

// A page that fails after the first fails the whole portal, so a partial
// listing is never passed off as complete.
func TestPortalPagesError(t *testing.T) {
	u := "https://jobs.example.org/api?page="
	d := &pageDoer{pages: map[string]string{u + "1": jobsJSON("1")}}
	p := pagedPortal(t, "listing_url: "+u+"{page}\npages: 3\n")
	_, err := (&portalSource{p: p, client: d}).Search(context.Background(), nil)
	if !errors.Is(err, ErrFetch) {
		t.Errorf("got %v, want ErrFetch", err)
	}
}

// {page} in listing_url without `pages` is refused, as is a negative
// `pages` (invalid/ holds the body and missing-placeholder cases).
func TestPortalPagesValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.yml")
	data := "name: X\nats: bespoke\njson_path: jobs\nfields: {title: t, url: u}\n" +
		"listing_url: https://jobs.example.org/api?page={page}\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPortal(path); err == nil || !strings.Contains(err.Error(), "listing_url: {page} is only allowed with pages") {
		t.Errorf("got %v", err)
	}
	if err := os.WriteFile(path, []byte(data+"pages: -1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPortal(path); err == nil || !strings.Contains(err.Error(), "pages: must be between 1 and 10") {
		t.Errorf("got %v", err)
	}
}
