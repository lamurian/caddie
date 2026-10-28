package doctor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const gateOn = "apply: {require_gate: true}\n"

// newWorkspace builds a complete workspace in a temp dir with the fixture
// history, the given preferences and any history file overrides.
func newWorkspace(t *testing.T, prefs string, history map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{"postings", "applied", "manual", "resumes", "cover-letters",
		"reviews", "history", "handoff"} {
		if err := os.Mkdir(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	src := "../../testdata/history"
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{"preferences.yml": prefs}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		files[filepath.Join("history", e.Name())] = string(data)
	}
	for name, content := range history {
		files[filepath.Join("history", name)] = content
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func run(root string, found bool) Report {
	return Run(Options{
		DotEnv: filepath.Join(root, "absent.env"),
		Lookup: func(k string) (string, bool) {
			if k == "CADDIE_WORKSPACE" {
				return root, true
			}
			return "", false
		},
		LookPath: func(f string) (string, error) {
			if found {
				return "/usr/bin/" + f, nil
			}
			return "", errors.New("not found")
		},
	})
}

func status(r Report, name string) Check {
	for _, c := range r.Checks {
		if c.Name == name {
			return c
		}
	}
	return Check{}
}

func TestHealthy(t *testing.T) {
	r := run(newWorkspace(t, gateOn, nil), true)
	if !r.OK {
		t.Fatalf("expected healthy: %+v", r.Checks)
	}
	for _, name := range []string{"config", "workspace", "preferences", "gate", "history", "quarto", "docker"} {
		if c := status(r, name); c.Status != StatusOK {
			t.Errorf("%s: %+v", name, c)
		}
	}
}

func TestMissingToolsAreWarnings(t *testing.T) {
	r := run(newWorkspace(t, gateOn, nil), false)
	if !r.OK || status(r, "quarto").Status != StatusWarn || status(r, "docker").Status != StatusWarn {
		t.Errorf("missing tools must warn, not fail: %+v", r.Checks)
	}
}

func TestGateDisabledFails(t *testing.T) {
	r := run(newWorkspace(t, "apply: {require_gate: false}\n", nil), true)
	c := status(r, "gate")
	if r.OK || c.Status != StatusFail || !strings.Contains(c.Message, "require_gate is false") {
		t.Errorf("got %+v", c)
	}
}

func TestMalformedHistoryFails(t *testing.T) {
	root := newWorkspace(t, gateOn, map[string]string{"employment.md": "## Employment\n\n### Analyst\nOrg\n"})
	c := status(run(root, true), "history")
	if c.Status != StatusFail || !strings.Contains(c.Message, "employment.md:3:") {
		t.Errorf("got %+v", c)
	}
}

func TestMissingDirsAndPreferences(t *testing.T) {
	root := newWorkspace(t, gateOn, nil)
	for _, p := range []string{"handoff", "preferences.yml"} {
		if err := os.Remove(filepath.Join(root, p)); err != nil {
			t.Fatal(err)
		}
	}
	r := run(root, true)
	if r.OK || !strings.Contains(status(r, "workspace").Message, "handoff") {
		t.Errorf("workspace: %+v", status(r, "workspace"))
	}
	if c := status(r, "preferences"); c.Status != StatusFail || !strings.Contains(c.Message, "preferences.example.yml") {
		t.Errorf("preferences: %+v", c)
	}
	if status(r, "gate").Status != StatusFail {
		t.Error("gate must fail when preferences do not load")
	}
}

func TestBadConfigStops(t *testing.T) {
	root := t.TempDir()
	dotenv := filepath.Join(root, ".env")
	if err := os.WriteFile(dotenv, []byte("CADDIE_BROWSER_MODE=remote\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := Run(Options{DotEnv: dotenv, Lookup: func(string) (string, bool) { return "", false }})
	if r.OK || len(r.Checks) != 1 || status(r, "config").Status != StatusFail {
		t.Errorf("got %+v", r)
	}
}
