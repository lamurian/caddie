package source

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const (
	greenhouseBadURL = "https://boards-api.greenhouse.io/v1/boards/nosuchco/jobs?content=true"
	leverBadURL      = "https://api.lever.co/v0/postings/nosuchco?mode=json"
)

var boardCases = []struct {
	name, bad, good, file, notice string
	src                           func(Doer) Source
}{
	{"greenhouse", greenhouseBadURL, greenhouseURL, "greenhouse-acme.json",
		"greenhouse-board: board nosuchco skipped: HTTP 404 — check the slug in sources.greenhouse_boards",
		func(c Doer) Source { return &greenhouse{client: c, boards: []string{"nosuchco", "acme"}} }},
	{"lever", leverBadURL, leverURL, "lever-globex.json",
		"lever-board: board nosuchco skipped: HTTP 404 — check the slug in sources.lever_boards",
		func(c Doer) Source { return &lever{client: c, boards: []string{"nosuchco", "globex"}} }},
}

// One unknown board (HTTP 404) is skipped with a notice; the good board's
// jobs are still returned and the source does not count as failed.
func TestBadBoardIsSkippedAlone(t *testing.T) {
	for _, c := range boardCases {
		t.Run(c.name, func(t *testing.T) {
			fake := newFake(t, map[string]string{c.good: c.file})
			res, err := Search(context.Background(), []Source{c.src(fake)}, nil, Filter{})
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Jobs) == 0 || len(fake.calls) != 2 || fake.calls[0] != c.bad {
				t.Errorf("%d jobs from %d calls, want the good board's jobs", len(res.Jobs), len(fake.calls))
			}
			if len(res.Notices) != 1 || res.Notices[0] != c.notice {
				t.Errorf("notices = %q, want %q", res.Notices, c.notice)
			}
		})
	}
}

// When every board fails the source fails, and with it a search of only it.
func TestEveryBoardFailingFailsSource(t *testing.T) {
	for _, c := range boardCases {
		t.Run(c.name, func(t *testing.T) {
			fake := newFake(t, map[string]string{c.good: "!err"})
			res, err := Search(context.Background(), []Source{c.src(fake)}, nil, Filter{})
			if !errors.Is(err, ErrFetch) || len(res.Jobs) != 0 {
				t.Fatalf("got %d jobs, err %v", len(res.Jobs), err)
			}
			for _, want := range []string{"every board failed", "board nosuchco skipped: HTTP 404",
				"connection refused"} {
				if len(res.Notices) != 1 || !strings.Contains(res.Notices[0], want) {
					t.Errorf("notices = %q, want %q", res.Notices, want)
				}
			}
		})
	}
}

// A board failure that is not a 404 keeps its redacted cause, no query.
func TestBoardErrorHidesQuery(t *testing.T) {
	fake := newFake(t, map[string]string{greenhouseURL: "!err"})
	g := &greenhouse{client: fake, boards: []string{"acme"}}
	_, err := g.Search(context.Background(), nil)
	if err == nil || strings.Contains(err.Error(), "content=true") || !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("got %v", err)
	}
}
