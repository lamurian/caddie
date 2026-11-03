package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/lamurian/caddie/internal/browse"
	"github.com/lamurian/caddie/internal/document"
)

// fakeBrowser records each argv; it prints out and returns err.
type fakeBrowser struct {
	calls [][]string
	out   string
	err   error
}

func (f *fakeBrowser) Run(_ context.Context, argv []string, stdout, _ io.Writer) error {
	f.calls = append(f.calls, argv)
	_, _ = io.WriteString(stdout, f.out)
	return f.err
}

// runBrowse runs the CLI in-process with b as the browser and no network.
func runBrowse(t *testing.T, vars map[string]string, b browse.Exec, args ...string) result {
	t.Helper()
	var out, errOut bytes.Buffer
	code := run(args, env{
		stdout: &out, stderr: &errOut,
		lookup:   func(k string) (string, bool) { v, ok := vars[k]; return v, ok },
		dotenv:   filepath.Join(t.TempDir(), "absent.env"),
		lookPath: func(string) (string, error) { return "", errors.New("not found") },
		client:   fakeHTTP(nil),
		today:    func() string { return "2026-10-28" },
		quarto:   func(string) document.Quarto { return noQuarto{t} },
		browser:  b,
	})
	return result{code, out.String(), errOut.String()}
}

func TestBrowsePassesArgsVerbatim(t *testing.T) {
	b := &fakeBrowser{out: "- button \"Apply\" [ref=e1]\n"}
	vars := map[string]string{"CADDIE_WORKSPACE": "/ws"}
	r := runBrowse(t, vars, b, "browse", "careers.who.int", "snapshot", "-i", "--json")
	if r.code != exitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	want := []string{"docker", "compose", "-f", "/ws/docker/compose.yml", "exec", "-T",
		"browser", "agent-browser", "--session", "careers.who.int", "snapshot", "-i", "--json"}
	if len(b.calls) != 1 || !slices.Equal(b.calls[0], want) {
		t.Errorf("argv %q, want %q", b.calls, want)
	}
	if r.stdout != b.out {
		t.Errorf("stdout %q, want agent-browser's output %q", r.stdout, b.out)
	}
}

func TestBrowseLocalMode(t *testing.T) {
	b := &fakeBrowser{}
	vars := map[string]string{"CADDIE_BROWSER_MODE": "local", "CADDIE_BROWSER_BIN": "/opt/ab"}
	r := runBrowse(t, vars, b, "browse", "inspira.un.org",
		"upload", "#cv", "/handoff/resume-x.pdf")
	want := []string{"/opt/ab", "--session", "inspira.un.org", "upload", "#cv", "/handoff/resume-x.pdf"}
	if r.code != exitOK || len(b.calls) != 1 || !slices.Equal(b.calls[0], want) {
		t.Errorf("exit %d, argv %q; want 0 and %q (%s)", r.code, b.calls, want, r.stderr)
	}
}

func TestBrowseErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
		err  error
		code int
		want string
	}{
		{"no args", nil, nil, exitUsage, "browse <session>"},
		{"help", []string{"--help"}, nil, exitUsage, "browse <session>"},
		{"no command", []string{"careers.who.int"}, nil, exitUsage, "no agent-browser command"},
		{"traversal session", []string{"../x", "open", "https://careers.who.int"}, nil, exitRefused, "invalid browse session"},
		{"spaced session", []string{"a b", "snapshot"}, nil, exitRefused, "invalid browse session"},
		{"upper session", []string{"Careers.who.int", "snapshot"}, nil, exitRefused, "invalid browse session"},
		{"session override", []string{"a.org", "--session=b.org", "snapshot"}, nil, exitRefused, "invalid browse session"},
		{"upload escape", []string{"a.org", "upload", "#cv", "/handoff/../etc/passwd"}, nil, exitRefused, "outside /handoff"},
		{"upload relative", []string{"a.org", "upload", "#cv", "handoff/cv.pdf"}, nil, exitRefused, "outside /handoff"},
		{"exec failure", []string{"a.org", "snapshot"}, errors.New("exit status 1"), exitExternal, "agent-browser failed"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := &fakeBrowser{err: c.err}
			r := runBrowse(t, nil, b, append([]string{"browse"}, c.args...)...)
			if r.code != c.code || !strings.Contains(r.stderr, c.want) {
				t.Errorf("exit %d, stderr %q; want %d and %q", r.code, r.stderr, c.code, c.want)
			}
			if ran := len(b.calls) > 0; ran != (c.err != nil) {
				t.Errorf("browser ran = %v, want %v", ran, c.err != nil)
			}
		})
	}
}

func TestBrowseTimeoutExit(t *testing.T) {
	vars := map[string]string{"CADDIE_BROWSE_TIMEOUT": "1ms"}
	r := runBrowse(t, vars, blockingBrowser{}, "browse", "a.org", "snapshot")
	if r.code != exitExternal || !strings.Contains(r.stderr, "timed out") {
		t.Errorf("exit %d, stderr %q; want %d and a timeout", r.code, r.stderr, exitExternal)
	}
}

// blockingBrowser runs until its context ends.
type blockingBrowser struct{}

func (blockingBrowser) Run(ctx context.Context, _ []string, _, _ io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("context never ended")
	}
}

// Every concrete `caddie browse` example in the skill passes validation.
func TestSkillBrowseExamplesValidate(t *testing.T) {
	data, err := os.ReadFile("../../.agents/skills/caddie/references/browse.md")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, line := range strings.Split(string(data), "\n") {
		rest, ok := strings.CutPrefix(line, "caddie browse ")
		if !ok || strings.Contains(rest, "<") {
			continue
		}
		f := strings.Fields(rest)
		if err := browse.ValidateSession(f[0]); err != nil {
			t.Error(err)
		}
		if err := browse.ValidateArgs(f[1:]); err != nil {
			t.Error(err)
		}
		n++
	}
	if n == 0 {
		t.Error("no concrete `caddie browse` example in browse.md")
	}
}
