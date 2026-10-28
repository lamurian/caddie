package memory

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/source"
)

func writeSeen(t *testing.T, content string) Store {
	t.Helper()
	s := Store{Dir: filepath.Join(t.TempDir(), ".caddie")}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.SeenPath(), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return s
}

// seen.jsonl tolerates hand-edited and blank lines and reports each one.
func TestReadLogToleratesHandEdits(t *testing.T) {
	s := writeSeen(t, `{"id":"adzuna:1","url":"u1","title":"A","company":"X","status":"shown","at":"2026-10-27"}

{"id":"adzuna:2", "status": "rejected"}
this line was typed by hand
{"url":"u9","status":"shown"}
{"id":"remotive:3","status":"maybe"}
{"id":"adzuna:1","status":"rejected","at":"2026-10-28"}`)
	l, err := s.Log()
	if err != nil {
		t.Fatal(err)
	}
	if got := len(l.Entries); got != 4 {
		t.Errorf("%d entries, want 4", got)
	}
	want := map[int]string{2: "blank line", 4: "not a JSON object", 5: "no id", 6: `unknown status "maybe"`}
	if len(l.Problems) != len(want) {
		t.Errorf("problems: %v", l.Problems)
	}
	for _, p := range l.Problems {
		if !strings.Contains(p.Message, want[p.Line]) || !strings.HasPrefix(p.String(), s.SeenPath()+":") {
			t.Errorf("line %d: %q, want %q", p.Line, p.String(), want[p.Line])
		}
	}
	latest := l.Latest()
	if latest[0].ID != "adzuna:1" || latest[0].Status != Rejected || len(latest) != 3 {
		t.Errorf("latest = %+v", latest)
	}
	if ids := l.IDs(); !ids["remotive:3"] || !ids["adzuna:2"] || ids["u9"] {
		t.Errorf("IDs = %v", ids)
	}
	if e, ok := l.ByURL("u1"); !ok || e.ID != "adzuna:1" {
		t.Errorf("ByURL = %+v %v", e, ok)
	}
}

func TestReadLogMissingFile(t *testing.T) {
	l, err := ReadLog(filepath.Join(t.TempDir(), "absent.jsonl"))
	if err != nil || len(l.Entries) != 0 || len(l.Problems) != 0 {
		t.Errorf("got %+v, %v", l, err)
	}
}

func TestMark(t *testing.T) {
	s := writeSeen(t, `{"id":"a:1","url":"u1","title":"T1","company":"C","status":"saved","at":"2026-10-01"}`)
	if err := s.Remember([]source.Job{{ID: "a:2", URL: "u2", Title: "T2", Company: "C2"}}); err != nil {
		t.Fatal(err)
	}
	res, err := s.Mark([]string{"a:1", "a:2", "zz:9", "a:2"}, Shown, "2026-10-28")
	if err != nil {
		t.Fatal(err)
	}
	want := MarkResult{Recorded: []string{"a:2", "zz:9"}, Unchanged: []string{"a:1", "a:2"},
		Unknown: []string{"zz:9"}, Problems: []string{}}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("got %+v, want %+v", res, want)
	}
	if res, _ = s.Mark([]string{"a:2"}, Rejected, "2026-10-29"); len(res.Recorded) != 1 {
		t.Errorf("rejected must replace shown: %+v", res)
	}
	l, _ := s.Log()
	got := l.Latest()
	wantLatest := []Entry{
		{ID: "a:1", URL: "u1", Title: "T1", Company: "C", Status: Saved, At: "2026-10-01"},
		{ID: "a:2", URL: "u2", Title: "T2", Company: "C2", Status: Rejected, At: "2026-10-29"},
		{ID: "zz:9", Status: Shown, At: "2026-10-28"},
	}
	if !reflect.DeepEqual(got, wantLatest) || len(l.Problems) != 0 {
		t.Errorf("latest:\n got %+v\nwant %+v\nproblems %v", got, wantLatest, l.Problems)
	}
}

func TestAppendRepairsMissingFinalNewline(t *testing.T) {
	s := writeSeen(t, `{"id":"a:1","status":"shown"}`)
	if err := s.Record(Entry{ID: "a:2", Status: Saved, At: "2026-10-28"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Record(Entry{ID: "a:2", Status: Shown, At: "2026-10-29"}); err != nil {
		t.Fatal(err)
	}
	l, _ := s.Log()
	if len(l.Entries) != 2 || len(l.Problems) != 0 {
		t.Errorf("entries %+v, problems %v", l.Entries, l.Problems)
	}
}

func TestRememberAndRecall(t *testing.T) {
	s := Store{Dir: filepath.Join(t.TempDir(), "state")}
	if _, ok, err := s.RecallURL("u1"); ok || err != nil {
		t.Fatalf("empty store: %v %v", ok, err)
	}
	_ = s.Remember([]source.Job{{ID: "a:1", URL: "u1", Title: "old"}, {ID: "a:2", URL: "u2"}})
	_ = s.Remember([]source.Job{{ID: "a:1", URL: "u1", Title: "new"}})
	jobs, err := s.Recall()
	if err != nil || len(jobs) != 2 || jobs["a:1"].Title != "new" {
		t.Errorf("recall = %+v, %v", jobs, err)
	}
	if j, ok, _ := s.RecallURL("u2"); !ok || j.ID != "a:2" {
		t.Errorf("RecallURL = %+v", j)
	}
}
