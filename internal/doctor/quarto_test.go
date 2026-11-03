package doctor

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// fakeQuarto records its one run and writes pdf to --output when non-nil.
type fakeQuarto struct {
	bin, dir, qmd string
	args          []string
	deadline      time.Time
	pdf           []byte
	out           string
	err           error
}

func (f *fakeQuarto) Run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	f.dir, f.args = dir, args
	f.deadline, _ = ctx.Deadline()
	if data, err := os.ReadFile(filepath.Join(dir, "doctor.qmd")); err == nil {
		f.qmd = string(data)
	}
	if i := slices.Index(args, "--output"); f.pdf != nil && i >= 0 {
		if err := os.WriteFile(filepath.Join(dir, args[i+1]), f.pdf, 0o600); err != nil {
			return nil, err
		}
	}
	return []byte(f.out), f.err
}

// quartoCheck runs doctor with q and returns the quarto line. It fails the
// test if the render left anything in the temp dir or used the workspace.
func quartoCheck(t *testing.T, found bool, q *fakeQuarto) (Check, Report) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	root := newWorkspace(t, gateOn, nil)
	r := Run(options(root, found, q))
	if left, _ := os.ReadDir(tmp); len(left) > 0 {
		t.Errorf("temp dir not removed: %v", left)
	}
	if q.dir != "" && (!strings.HasPrefix(q.dir, tmp) || strings.HasPrefix(q.dir, root)) {
		t.Errorf("rendered in %s, want a fresh dir under %s", q.dir, tmp)
	}
	return status(r, "quarto"), r
}

func TestQuartoMissingWarns(t *testing.T) {
	q := &fakeQuarto{}
	c, r := quartoCheck(t, false, q)
	if !r.OK || c.Status != StatusWarn || !strings.Contains(c.Message, "needed by `caddie render`") {
		t.Errorf("got %+v", c)
	}
	if q.args != nil {
		t.Errorf("quarto ran although missing: %q", q.args)
	}
}

func TestQuartoRendersTestPDF(t *testing.T) {
	q := &fakeQuarto{pdf: []byte("%PDF-fake")}
	c, r := quartoCheck(t, true, q)
	if !r.OK || c.Status != StatusOK || c.Message != "found at /usr/bin/quarto; rendered a test PDF" {
		t.Errorf("got %+v", c)
	}
	want := []string{"render", "doctor.qmd", "--to", "typst", "--output", "doctor.pdf"}
	if q.bin != "/usr/bin/quarto" || !slices.Equal(q.args, want) {
		t.Errorf("ran %s %q, want /usr/bin/quarto %q", q.bin, q.args, want)
	}
	if !strings.HasPrefix(q.qmd, "---\nformat: typst\n---\n") {
		t.Errorf("doctor.qmd = %q", q.qmd)
	}
	if q.deadline.IsZero() || time.Until(q.deadline) > quartoTimeout {
		t.Errorf("render must run under a %s timeout, deadline %v", quartoTimeout, q.deadline)
	}
}

func TestQuartoErrorFails(t *testing.T) {
	var out []string
	for i := 1; i <= 9; i++ {
		out = append(out, "line "+string(rune('0'+i)))
	}
	out = append(out, "", "ERROR: typst compile failed", "")
	q := &fakeQuarto{out: strings.Join(out, "\n"), err: errors.New("exit status 1")}
	c, r := quartoCheck(t, true, q)
	if r.OK || c.Status != StatusFail {
		t.Fatalf("a broken quarto must fail: %+v", c)
	}
	for _, want := range []string{"cannot render", "exit status 1", "line 6 | line 7 | line 8 | line 9 | ERROR: typst compile failed"} {
		if !strings.Contains(c.Message, want) {
			t.Errorf("message lacks %q: %s", want, c.Message)
		}
	}
	if strings.Contains(c.Message, "line 5") || strings.Contains(c.Message, "\n") {
		t.Errorf("message should be one line with only the output tail: %q", c.Message)
	}
}

func TestQuartoWithoutPDFFails(t *testing.T) {
	for name, pdf := range map[string][]byte{"no file": nil, "empty file": {}} {
		t.Run(name, func(t *testing.T) {
			c, r := quartoCheck(t, true, &fakeQuarto{pdf: pdf, out: "Output created: doctor.pdf"})
			if r.OK || c.Status != StatusFail || !strings.Contains(c.Message, "wrote no doctor.pdf: Output created") {
				t.Errorf("got %+v", c)
			}
		})
	}
}

// TestQuartoReal renders with the real quarto; skipped when it is absent.
func TestQuartoReal(t *testing.T) {
	path, err := exec.LookPath("quarto")
	if err != nil {
		t.Skip("quarto not on PATH")
	}
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	opts := options(newWorkspace(t, gateOn, nil), true, nil)
	opts.LookPath, opts.Quarto = exec.LookPath, nil
	c := status(Run(opts), "quarto")
	if c.Status != StatusOK || c.Message != "found at "+path+"; rendered a test PDF" {
		t.Errorf("got %+v", c)
	}
	if left, _ := os.ReadDir(tmp); len(left) > 0 {
		t.Errorf("temp dir not removed: %v", left)
	}
}
