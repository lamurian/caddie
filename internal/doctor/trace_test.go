package doctor

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// runTrace runs doctor on a workspace whose .caddie/trace is replaced by
// setup (given its path), in the given browser mode.
func runTrace(t *testing.T, mode string, setup func(t *testing.T, dir string)) Check {
	t.Helper()
	root := newWorkspace(t, gateOn, nil)
	dir := filepath.Join(root, ".caddie", "trace")
	if err := os.Remove(dir); err != nil {
		t.Fatal(err)
	}
	setup(t, dir)
	opts := options(root, true, &fakeQuarto{pdf: []byte("%PDF-fake")})
	opts.Lookup = func(k string) (string, bool) {
		switch k {
		case "CADDIE_WORKSPACE":
			return root, true
		case "CADDIE_BROWSER_MODE":
			return mode, true
		}
		return "", false
	}
	return status(Run(opts), "trace")
}

func mkdirMode(perm os.FileMode) func(t *testing.T, dir string) {
	return func(t *testing.T, dir string) {
		t.Helper()
		if err := os.Mkdir(dir, perm); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(dir, perm); err != nil { // umask-proof
			t.Fatal(err)
		}
	}
}

func TestTrace(t *testing.T) {
	none := func(*testing.T, string) {}
	file := func(t *testing.T, dir string) {
		t.Helper()
		if err := os.WriteFile(dir, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cases := []struct {
		name, mode string
		setup      func(*testing.T, string)
		want       Status
		msg        string
	}{
		{"missing compose", "compose", none, StatusFail, "refuses to start without it: mkdir -p .caddie/trace"},
		{"missing local", "local", none, StatusWarn, "archives no trace: mkdir -p .caddie/trace"},
		{"not a directory", "compose", file, StatusFail, ".caddie/trace is not a directory"},
		{"writable", "compose", mkdirMode(0o777), StatusOK, ".caddie/trace present"},
		{"owner only, local", "local", mkdirMode(0o700), StatusOK, "present"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := runTrace(t, c.mode, c.setup)
			if got.Status != c.want || !strings.Contains(got.Message, c.msg) {
				t.Errorf("trace = %+v, want %s containing %q", got, c.want, c.msg)
			}
		})
	}
}

func TestTraceOwnerOnly(t *testing.T) {
	got := runTrace(t, "compose", mkdirMode(0o700))
	want := StatusOK
	if runtime.GOOS == "linux" && os.Getuid() != browserUID {
		want = StatusWarn
	}
	if got.Status != want || (want == StatusWarn && !strings.Contains(got.Message, "setfacl -m u:1000:rwx")) {
		t.Errorf("trace = %+v, want %s", got, want)
	}
}

func TestTraceMissingFailsReport(t *testing.T) {
	root := newWorkspace(t, gateOn, nil)
	if err := os.Remove(filepath.Join(root, ".caddie", "trace")); err != nil {
		t.Fatal(err)
	}
	if r := run(root, true); r.OK {
		t.Errorf("a missing trace dir must fail in compose mode: %+v", r.Checks)
	}
}

func TestRelTo(t *testing.T) {
	cases := map[string]string{"/w/.caddie/trace": ".caddie/trace", "/srv/state/trace": "/srv/state/trace"}
	for path, want := range cases {
		if got := relTo("/w", filepath.FromSlash(path)); got != filepath.FromSlash(want) {
			t.Errorf("relTo(%q) = %q, want %q", path, got, want)
		}
	}
}
