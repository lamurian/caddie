package posting

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/lamurian/caddie/internal/workspace"
)

func merge(t *testing.T, body string, steps ...string) (string, Checklist) {
	t.Helper()
	out, list, err := MergeManualSteps([]byte(body), steps)
	if err != nil {
		t.Fatalf("MergeManualSteps: %v", err)
	}
	return string(out), list
}

func TestMergeManualSteps(t *testing.T) {
	cases := []struct {
		name, body string
		steps      []string
		want       string
		added      int
	}{
		{"new section", "## Description\n\nText.\n\n\n", []string{" Email  the\trecruiter ", "Post CV"},
			"## Description\n\nText.\n\n## Manual steps\n\n- [ ] Email the recruiter\n- [ ] Post CV\n", 2},
		{"empty body", "", []string{"a"}, "## Manual steps\n\n- [ ] a\n", 1},
		{"no trailing newline", "## Description\n\nText.", []string{"a"},
			"## Description\n\nText.\n\n## Manual steps\n\n- [ ] a\n", 1},
		{"merge after last item", "## Manual steps\n\nIntro.\n\n- [x] Call HR\n- [ ] Fill form\n\nNote.\n", []string{"call  hr", "Upload ID"},
			"## Manual steps\n\nIntro.\n\n- [x] Call HR\n- [ ] Fill form\n- [ ] Upload ID\n\nNote.\n", 1},
		{"capital X kept ticked", "## Manual steps\n\n- [X] Call HR\n", []string{"Call HR"},
			"## Manual steps\n\n- [X] Call HR\n", 0},
		{"following heading kept", "## Manual steps\n\n- [ ] a\n\n## Notes\n\n- [ ] b  \n", []string{"b"},
			"## Manual steps\n\n- [ ] a\n- [ ] b\n\n## Notes\n\n- [ ] b  \n", 1},
		{"bare heading", "## Manual steps\n\n", []string{"a"}, "## Manual steps\n\n- [ ] a\n", 1},
		{"bare heading at EOF", "Text.\n\n## Manual steps", []string{"a"}, "Text.\n\n## Manual steps\n\n- [ ] a\n", 1},
		{"text but no items", "## Manual steps\nSee below.\n\n## Reviews\n", []string{"a"},
			"## Manual steps\nSee below.\n\n- [ ] a\n\n## Reviews\n", 1},
		{"item at EOF without newline", "## Manual steps\n\n- [ ] a", []string{"b"}, "## Manual steps\n\n- [ ] a\n- [ ] b\n", 1},
		{"CRLF new section", "## Description\r\n\r\nText.\r\n", []string{"a"},
			"## Description\r\n\r\nText.\r\n\r\n## Manual steps\r\n\r\n- [ ] a\r\n", 1},
		{"CRLF merge", "## Manual steps\r\n\r\n- [x] a\r\n", []string{"A", "b"},
			"## Manual steps\r\n\r\n- [x] a\r\n- [ ] b\r\n", 1},
		{"input repeats", "", []string{"a", " A "}, "## Manual steps\n\n- [ ] a\n", 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, list := merge(t, c.body, c.steps...)
			if got != c.want {
				t.Errorf("body:\n got %q\nwant %q", got, c.want)
			}
			if list.Added != c.added {
				t.Errorf("added = %d, want %d", list.Added, c.added)
			}
			again, list2 := merge(t, got, c.steps...)
			if again != got || list2.Added != 0 {
				t.Errorf("second run changed body (added %d):\n%q", list2.Added, again)
			}
		})
	}
}

func TestMergeManualStepsChecklist(t *testing.T) {
	_, list := merge(t, "## Manual steps\n\n- [x] Call HR  \n- [ ] Fill form\n", "call hr", "Upload ID")
	want := []Step{{"Call HR", true}, {"Fill form", false}, {"Upload ID", false}}
	if len(list.Steps) != len(want) {
		t.Fatalf("steps = %+v", list.Steps)
	}
	for i := range want {
		if list.Steps[i] != want[i] {
			t.Errorf("step %d = %+v, want %+v", i, list.Steps[i], want[i])
		}
	}
}

func TestMergeManualStepsErrors(t *testing.T) {
	if _, _, err := MergeManualSteps(nil, nil); !errors.Is(err, ErrNoSteps) {
		t.Errorf("no steps: err = %v", err)
	}
	if _, _, err := MergeManualSteps(nil, []string{"a", " \t "}); !errors.Is(err, ErrEmptyStep) {
		t.Errorf("empty step: err = %v", err)
	}
}

const manualFront = "---\nslug: acme\ntitle:   Data Scientist   # kept as written\nstatus: manual\nmanual_reason: portal needs an account\n---\n"

func TestSetManualSteps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "acme.md")
	if err := os.WriteFile(path, []byte(manualFront+"## Description\n\nText.\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	list, err := SetManualSteps(path, []string{"Create an account"})
	if err != nil {
		t.Fatal(err)
	}
	if list.Path != path || list.Added != 1 || len(list.Steps) != 1 {
		t.Errorf("checklist = %+v", list)
	}
	got, _ := os.ReadFile(path)
	want := manualFront + "## Description\n\nText.\n\n## Manual steps\n\n- [ ] Create an account\n"
	if string(got) != want {
		t.Errorf("file:\n got %q\nwant %q", got, want)
	}
	if fi, _ := os.Stat(path); fi.Mode().Perm() != 0o640 {
		t.Errorf("mode = %v", fi.Mode().Perm())
	}
	before, _ := os.Stat(path)
	list, err = SetManualSteps(path, []string{"create an ACCOUNT"})
	if err != nil || list.Added != 0 {
		t.Fatalf("second run: %+v, %v", list, err)
	}
	after, _ := os.Stat(path)
	if !os.SameFile(before, after) {
		t.Error("second run rewrote the file")
	}
}

func TestSetManualStepsErrors(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.md")
	if err := os.WriteFile(bad, []byte("no frontmatter\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := SetManualSteps(bad, []string{"a"}); !errors.Is(err, workspace.ErrBadPosting) {
		t.Errorf("bad posting: err = %v", err)
	}
	if _, err := SetManualSteps(filepath.Join(dir, "missing.md"), []string{"a"}); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("missing: err = %v", err)
	}
	good := filepath.Join(dir, "good.md")
	if err := os.WriteFile(good, []byte(manualFront), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := SetManualSteps(good, []string{""}); !errors.Is(err, ErrEmptyStep) {
		t.Errorf("empty step: err = %v", err)
	}
	if got, _ := os.ReadFile(good); string(got) != manualFront {
		t.Error("failed call changed the file")
	}
}
