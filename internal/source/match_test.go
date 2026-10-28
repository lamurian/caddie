package source

import (
	"context"
	"reflect"
	"testing"
)

// boilerplateTerms all occur in the "About us" text every job on the
// Initrode and Initech boards repeats.
var boilerplateTerms = []string{"digital health", "health informatics"}

// On a company board only the title and the department or team are
// matched, so boilerplate in every description does not match every job.
func TestBoardsIgnoreBoilerplateDescriptions(t *testing.T) {
	fake := newFake(t, map[string]string{
		"https://boards-api.greenhouse.io/v1/boards/initrode/jobs?content=true": "greenhouse-initrode.json",
		"https://api.lever.co/v0/postings/initech?mode=json":                    "lever-initech.json",
	})
	cases := []struct {
		src  Source
		want []string
	}{
		// 902 by title, 903 by department; 901 and 904 only by boilerplate.
		{&greenhouse{client: fake, boards: []string{"initrode"}},
			[]string{"greenhouse-board:902", "greenhouse-board:903"}},
		// The analyst by team; the counsel only by boilerplate.
		{&lever{client: fake, boards: []string{"initech"}},
			[]string{"lever-board:aaaaaaaa-1111-4222-8333-000000000002"}},
	}
	for _, c := range cases {
		jobs, err := c.src.Search(context.Background(), boilerplateTerms)
		if err != nil {
			t.Fatal(err)
		}
		if got := ids(jobs); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: %v, want %v", c.src.Name(), got, c.want)
		}
	}
}

func board(id, company string) Job {
	return Job{ID: id, URL: id, Source: Greenhouse, Title: "Analyst", Company: company}
}

// --limit is shared round-robin across boards and sources, so a long first
// board cannot crowd out later ones; kept jobs stay in their order.
func TestLimitIsSharedAcrossFeeds(t *testing.T) {
	srcs := []Source{
		stubSource{name: "a", jobs: []Job{{ID: "a:1", URL: "a1", Source: "a"}, {ID: "a:2", URL: "a2", Source: "a"}}},
		stubSource{name: Greenhouse, jobs: []Job{board("g:1", "Hooli"), board("g:2", "Hooli"),
			board("g:3", "Hooli"), board("g:4", "Vandelay")}},
	}
	cases := map[int][]string{
		3: {"a:1", "g:1", "g:4"},
		4: {"a:1", "a:2", "g:1", "g:4"},
		5: {"a:1", "a:2", "g:1", "g:2", "g:4"},
		0: {"a:1", "a:2", "g:1", "g:2", "g:3", "g:4"},
	}
	for n, want := range cases {
		res, err := Search(context.Background(), srcs, nil, Filter{Limit: n})
		if err != nil {
			t.Fatal(err)
		}
		if got := ids(res.Jobs); !reflect.DeepEqual(got, want) {
			t.Errorf("limit %d: %v, want %v", n, got, want)
		}
	}
}
