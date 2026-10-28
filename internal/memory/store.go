package memory

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/lamurian/caddie/internal/source"
)

// File names inside the state directory.
const (
	SeenFile       = "seen.jsonl"
	LastSearchFile = "last-search.json"
)

// maxRemembered caps how many jobs last-search.json keeps across searches.
const maxRemembered = 1000

// Store is Caddie's memory in the state directory (`.caddie/`).
type Store struct{ Dir string }

// SeenPath is the path of seen.jsonl.
func (s Store) SeenPath() string { return filepath.Join(s.Dir, SeenFile) }

// Log reads seen.jsonl (see ReadLog).
func (s Store) Log() (Log, error) { return ReadLog(s.SeenPath()) }

// MarkResult reports what `caddie seen add` did with each ID.
type MarkResult struct {
	Recorded  []string `json:"recorded"`
	Unchanged []string `json:"unchanged"`
	Unknown   []string `json:"unknown"`
	Problems  []string `json:"problems"`
}

// rank orders statuses: a weaker status never replaces a stronger one, so
// showing a job again cannot undo its rejection or its saving.
var rank = map[Status]int{Shown: 1, Rejected: 2, Saved: 2}

// Mark records status for each ID at the given date (YYYY-MM-DD). URL,
// title and company come from the last search, or from an earlier entry;
// an ID found in neither is still recorded and listed as Unknown.
func (s Store) Mark(ids []string, status Status, at string) (MarkResult, error) {
	res := MarkResult{Recorded: []string{}, Unchanged: []string{}, Unknown: []string{}, Problems: []string{}}
	log, err := s.Log()
	if err != nil {
		return res, err
	}
	for _, p := range log.Problems {
		res.Problems = append(res.Problems, p.String())
	}
	jobs, err := s.Recall()
	if err != nil {
		return res, err
	}
	latest := map[string]Entry{}
	for _, e := range log.Latest() {
		latest[e.ID] = e
	}
	var add []Entry
	for _, id := range ids {
		prev, seen := latest[id]
		if seen && rank[prev.Status] >= rank[status] {
			res.Unchanged = append(res.Unchanged, id)
			continue
		}
		e := Entry{ID: id, URL: prev.URL, Title: prev.Title, Company: prev.Company, Status: status, At: at}
		if j, ok := jobs[id]; ok {
			e.URL, e.Title, e.Company = j.URL, j.Title, j.Company
		} else if !seen {
			res.Unknown = append(res.Unknown, id)
		}
		latest[id] = e
		add = append(add, e)
		res.Recorded = append(res.Recorded, id)
	}
	return res, appendEntries(s.SeenPath(), add)
}

// Record appends e unless its ID already has an equal or stronger status.
func (s Store) Record(e Entry) error {
	log, err := s.Log()
	if err != nil {
		return err
	}
	for _, prev := range log.Latest() {
		if prev.ID == e.ID && rank[prev.Status] >= rank[e.Status] {
			return nil
		}
	}
	return appendEntries(s.SeenPath(), []Entry{e})
}

type lastSearch struct {
	Jobs []source.Job `json:"jobs"`
}

// Remember stores the jobs of a search, newest first, so that `seen add`
// and `posting save` can recover a job's details from its ID or URL.
func (s Store) Remember(jobs []source.Job) error {
	old, err := s.recallList()
	if err != nil {
		return err
	}
	have := map[string]bool{}
	for _, j := range jobs {
		have[j.ID] = true
	}
	for _, j := range old {
		if !have[j.ID] && len(jobs) < maxRemembered {
			jobs = append(jobs, j)
		}
	}
	data, err := json.MarshalIndent(lastSearch{Jobs: jobs}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return err
	}
	tmp := filepath.Join(s.Dir, LastSearchFile+".tmp")
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(s.Dir, LastSearchFile))
}

// Recall returns the remembered jobs by ID.
func (s Store) Recall() (map[string]source.Job, error) {
	list, err := s.recallList()
	out := make(map[string]source.Job, len(list))
	for _, j := range list {
		out[j.ID] = j
	}
	return out, err
}

// RecallURL returns the remembered job with this URL.
func (s Store) RecallURL(url string) (source.Job, bool, error) {
	list, err := s.recallList()
	for _, j := range list {
		if j.URL == url {
			return j, true, err
		}
	}
	return source.Job{}, false, err
}

// recallList reads last-search.json; a missing or unreadable file is
// treated as empty because it is only a cache.
func (s Store) recallList() ([]source.Job, error) {
	data, err := os.ReadFile(filepath.Join(s.Dir, LastSearchFile))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ls lastSearch
	if json.Unmarshal(data, &ls) != nil {
		return nil, nil
	}
	return ls.Jobs, nil
}
