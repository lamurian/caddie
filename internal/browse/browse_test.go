package browse

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"testing"
	"time"

	"github.com/lamurian/caddie/internal/config"
)

// fakeExec records each call and the deadline it saw.
type fakeExec struct {
	calls       [][]string
	deadline    time.Time
	hasDeadline bool
	err         error
	block       bool // wait for ctx to end, like a hung browser
}

func (f *fakeExec) Run(ctx context.Context, argv []string, stdout, _ io.Writer) error {
	f.calls = append(f.calls, argv)
	f.deadline, f.hasDeadline = ctx.Deadline()
	if f.block {
		<-ctx.Done()
		return ctx.Err()
	}
	_, _ = io.WriteString(stdout, "ok\n")
	return f.err
}

func testConfig(mode string) config.Config {
	return config.Config{
		Workspace:      "/ws",
		BrowserMode:    mode,
		BrowserBin:     "/opt/agent-browser",
		ComposeFile:    "docker/compose.yml",
		ComposeService: "browser",
		BrowseTimeout:  30 * time.Second,
	}
}

func TestBrowseArgv(t *testing.T) {
	args := []string{"upload", "@e3", "/handoff/Jane-resume.pdf"}
	tests := []struct {
		mode string
		want []string
	}{
		{"compose", []string{"docker", "compose", "-f", "/ws/docker/compose.yml",
			"exec", "-T", "browser", "agent-browser",
			"--session", "careers.who.int", "upload", "@e3", "/handoff/Jane-resume.pdf"}},
		{"local", []string{"/opt/agent-browser",
			"--session", "careers.who.int", "upload", "@e3", "/handoff/Jane-resume.pdf"}},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			ex := &fakeExec{}
			var out bytes.Buffer
			err := Browse(context.Background(), testConfig(tt.mode), ex,
				"careers.who.int", args, &out, io.Discard)
			if err != nil {
				t.Fatalf("Browse: %v", err)
			}
			if len(ex.calls) != 1 || !slices.Equal(ex.calls[0], tt.want) {
				t.Fatalf("argv = %q, want %q", ex.calls, tt.want)
			}
			if out.String() != "ok\n" {
				t.Errorf("stdout = %q", out.String())
			}
		})
	}
}

func TestArgvBadMode(t *testing.T) {
	_, err := Argv(testConfig("ssh"), "a", nil)
	if !errors.Is(err, config.ErrConfig) {
		t.Fatalf("err = %v, want ErrConfig", err)
	}
}

func TestBrowseRejectsBeforeExec(t *testing.T) {
	tests := []struct {
		name    string
		session string
		args    []string
		want    error
	}{
		{"traversal session", "../x", []string{"open", "https://a.org"}, ErrInvalidSession},
		{"space session", "a b", []string{"snapshot"}, ErrInvalidSession},
		{"empty session", "", []string{"snapshot"}, ErrInvalidSession},
		{"session override", "a.org", []string{"--session", "b", "open"}, ErrInvalidSession},
		{"upload traversal", "a.org", []string{"upload", "@e1", "/handoff/../etc/passwd"},
			ErrUploadOutsideHandoff},
		{"upload relative", "a.org", []string{"upload", "@e1", "handoff/a.pdf"},
			ErrUploadOutsideHandoff},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ex := &fakeExec{}
			err := Browse(context.Background(), testConfig("compose"), ex,
				tt.session, tt.args, io.Discard, io.Discard)
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if len(ex.calls) != 0 {
				t.Fatalf("Exec called with %q", ex.calls)
			}
		})
	}
}

func TestBrowseAppliesTimeout(t *testing.T) {
	ex := &fakeExec{}
	cfg := testConfig("local")
	start := time.Now()
	if err := Browse(context.Background(), cfg, ex, "a.org", []string{"snapshot"},
		io.Discard, io.Discard); err != nil {
		t.Fatalf("Browse: %v", err)
	}
	if !ex.hasDeadline {
		t.Fatal("Exec saw no deadline")
	}
	if d := ex.deadline.Sub(start); d <= 0 || d > cfg.BrowseTimeout+time.Second {
		t.Fatalf("deadline in %s, want about %s", d, cfg.BrowseTimeout)
	}
}

func TestBrowseTimeoutError(t *testing.T) {
	ex := &fakeExec{block: true}
	cfg := testConfig("local")
	cfg.BrowseTimeout = 10 * time.Millisecond
	err := Browse(context.Background(), cfg, ex, "a.org", []string{"wait", "5000"},
		io.Discard, io.Discard)
	if !errors.Is(err, ErrTimeout) || !errors.Is(err, ErrExec) {
		t.Fatalf("err = %v, want ErrTimeout and ErrExec", err)
	}
}

func TestBrowseExecError(t *testing.T) {
	ex := &fakeExec{err: errors.New("exit status 1")}
	err := Browse(context.Background(), testConfig("local"), ex, "a.org",
		[]string{"click", "@e9"}, io.Discard, io.Discard)
	if !errors.Is(err, ErrExec) || errors.Is(err, ErrTimeout) {
		t.Fatalf("err = %v, want ErrExec only", err)
	}
}

func TestOSExec(t *testing.T) {
	var out bytes.Buffer
	err := OSExec{}.Run(context.Background(), []string{"sh", "-c", "echo hi"}, &out, io.Discard)
	if err != nil {
		t.Skipf("sh unavailable: %v", err)
	}
	if out.String() != "hi\n" {
		t.Fatalf("stdout = %q", out.String())
	}
	if err := (OSExec{}).Run(context.Background(), nil, &out, &out); err == nil {
		t.Fatal("empty argv: want error")
	}
}
