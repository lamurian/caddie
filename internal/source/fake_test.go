package source

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

const fixtures = "../../testdata/sources"

// fakeDoer replays recorded responses: routes maps a full URL to a file in
// testdata/sources/. Any other URL is a 404; "!err" fails the request.
type fakeDoer struct {
	t      *testing.T
	routes map[string]string
	calls  []string
}

func (f *fakeDoer) Do(req *http.Request) (*http.Response, error) {
	u := req.URL.String()
	f.calls = append(f.calls, u)
	file, ok := f.routes[u]
	if file == "!err" {
		return nil, errors.New("connection refused")
	}
	if !ok {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewReader(nil))}, nil
	}
	data, err := os.ReadFile(filepath.Join(fixtures, file))
	if err != nil {
		f.t.Fatal(err)
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(data))}, nil
}

func newFake(t *testing.T, routes map[string]string) *fakeDoer {
	return &fakeDoer{t: t, routes: routes}
}

// withoutSummary blanks summaries so expected jobs stay readable; every
// summary must still be non-empty.
func withoutSummary(t *testing.T, jobs []Job) []Job {
	t.Helper()
	out := make([]Job, len(jobs))
	for i, j := range jobs {
		if j.Summary == "" {
			t.Errorf("%s: empty summary", j.ID)
		}
		j.Summary = ""
		out[i] = j
	}
	return out
}
