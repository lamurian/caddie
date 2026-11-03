package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/document"
)

// workspaceWith builds a complete temp workspace holding the fixture
// history (plus overrides) and the given preferences.
func workspaceWith(t *testing.T, prefs string, history map[string]string) map[string]string {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{"postings", "applied", "manual", "resumes", "cover-letters",
		"reviews", "history", "handoff"} {
		if err := os.Mkdir(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	src, err := filepath.Glob("../../testdata/history/*.md")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{"preferences.yml": prefs}
	for _, f := range src {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		files[filepath.Join("history", filepath.Base(f))] = string(data)
	}
	for name, content := range history {
		files[filepath.Join("history", name)] = content
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return map[string]string{"CADDIE_WORKSPACE": root}
}

func TestDoctorHealthy(t *testing.T) {
	r := runCaddie(t, workspaceWith(t, "apply: {require_gate: true}\n", nil), "doctor")
	if r.code != exitOK || !strings.Contains(r.stdout, "warn  quarto") {
		t.Errorf("got %+v", r)
	}
}

func TestDoctorGateDisabled(t *testing.T) {
	r := runCaddie(t, workspaceWith(t, "apply: {require_gate: false}\n", nil), "doctor")
	if r.code != exitCheckFailed || !strings.Contains(r.stdout, "fail  gate") ||
		!strings.Contains(r.stdout, "apply.require_gate is false") {
		t.Errorf("got %+v", r)
	}
	if !strings.Contains(r.stderr, "1 problem(s) found") {
		t.Errorf("stderr = %q", r.stderr)
	}
}

func TestDoctorMalformedHistory(t *testing.T) {
	vars := workspaceWith(t, "apply: {require_gate: true}\n",
		map[string]string{"teaching.md": "## Teaching\n### TA\nUniversity\nsometime\n"})
	r := runCaddie(t, vars, "doctor", "--json")
	if r.code != exitCheckFailed {
		t.Fatalf("exit %d, want %d: %+v", r.code, exitCheckFailed, r)
	}
	var rep struct {
		OK     bool
		Checks []struct{ Name, Status, Message string }
	}
	if err := json.Unmarshal([]byte(r.stdout), &rep); err != nil {
		t.Fatalf("doctor --json is not JSON: %v\n%s", err, r.stdout)
	}
	for _, c := range rep.Checks {
		if c.Name == "history" && c.Status == "fail" && strings.Contains(c.Message, "teaching.md:4: bad date range") {
			return
		}
	}
	t.Errorf("no history failure naming teaching.md:4 in %+v", rep)
}

// doctorWithQuarto runs doctor with every tool found and q as quarto.
func doctorWithQuarto(t *testing.T, q document.Quarto) result {
	t.Helper()
	vars := workspaceWith(t, "apply: {require_gate: true}\n", nil)
	var out, errOut bytes.Buffer
	code := run([]string{"doctor"}, env{
		stdout: &out, stderr: &errOut,
		lookup:   func(k string) (string, bool) { v, ok := vars[k]; return v, ok },
		dotenv:   filepath.Join(t.TempDir(), "absent.env"),
		lookPath: func(f string) (string, error) { return "/opt/bin/" + f, nil },
		quarto: func(bin string) document.Quarto {
			if bin != "/opt/bin/quarto" {
				t.Errorf("quarto bin %q", bin)
			}
			return q
		},
	})
	return result{code, out.String(), errOut.String()}
}

func TestDoctorQuartoRenders(t *testing.T) {
	q := &fakeQuarto{}
	r := doctorWithQuarto(t, q)
	if r.code != exitOK || !strings.Contains(r.stdout, "ok    quarto       found at /opt/bin/quarto; rendered a test PDF") {
		t.Errorf("got %+v", r)
	}
	if len(q.calls) != 1 || q.calls[0][1] != "doctor.qmd" {
		t.Errorf("quarto calls %q", q.calls)
	}
}

func TestDoctorQuartoBroken(t *testing.T) {
	r := doctorWithQuarto(t, &fakeQuarto{fail: true})
	if r.code != exitCheckFailed || !strings.Contains(r.stdout, "fail  quarto") ||
		!strings.Contains(r.stdout, "ERROR: compile failed") {
		t.Errorf("got %+v", r)
	}
}
