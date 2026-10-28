package source

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// partial is returned, alongside the jobs, by a source with several boards
// when some of them failed. Search turns each notice into a result notice
// and keeps the jobs; the source does not count as failed.
type partial struct{ notices []string }

func (p *partial) Error() string { return strings.Join(p.notices, "; ") }

// searchBoards fetches each board in turn. A failing board is skipped with
// a notice naming it; only when every board fails is the result an error.
// key is the preferences.yml list the board slugs come from.
func searchBoards(name, key string, boards []string, fetch func(board string) ([]Job, error)) ([]Job, error) {
	var jobs []Job
	var reasons []string
	for _, board := range boards {
		js, err := fetch(board)
		if err != nil {
			reasons = append(reasons, "board "+board+" skipped: "+boardReason(err, key))
			continue
		}
		jobs = append(jobs, js...)
	}
	switch {
	case len(boards) > 0 && len(reasons) == len(boards):
		return nil, fmt.Errorf("%w: every board failed: %s", ErrFetch, strings.Join(reasons, "; "))
	case len(reasons) > 0:
		notices := make([]string, len(reasons))
		for i, r := range reasons {
			notices[i] = name + ": " + r
		}
		return jobs, &partial{notices: notices}
	}
	return jobs, nil
}

// boardReason explains a board failure; a 404 almost always means a wrong
// slug. The error text never holds a query string (see get).
func boardReason(err error, key string) string {
	var s statusError
	if errors.As(err, &s) && int(s) == http.StatusNotFound {
		return s.Error() + " — check the slug in " + key
	}
	return err.Error()
}
