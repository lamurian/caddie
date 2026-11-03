package doctor

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

// reply is what the fake docker prints for one call.
type reply struct {
	out string
	err error
}

// fakeDocker answers `docker compose -f <file> <args…>` from replies keyed
// by the args after the file; unknown calls fail. It records every argv.
type fakeDocker struct {
	replies map[string]reply
	calls   [][]string
}

func (f *fakeDocker) Run(_ context.Context, argv []string, stdout, _ io.Writer) error {
	f.calls = append(f.calls, argv)
	if len(argv) < 4 {
		return errors.New("short argv")
	}
	rep, ok := f.replies[strings.Join(argv[4:], " ")]
	if !ok {
		return errors.New("unexpected call: " + strings.Join(argv, " "))
	}
	_, _ = io.WriteString(stdout, rep.out)
	return rep.err
}

const (
	cfgKey   = "config --format json"
	psKey    = "ps --format json"
	curlArgs = "exec -T browser curl -sS -o /dev/null --max-time 30 -w %{http_connect} "
	denyKey  = curlArgs + deniedURL
	allowKey = curlArgs + allowedURL

	cfgJSON   = `{"services":{"browser":{"image":"b"},"proxy":{"image":"p"}}}`
	browserUp = `{"Service":"browser","State":"running","Publishers":[]}`
	proxyUp   = `{"Service":"proxy","State":"running","Publishers":[]}`
)

var errExit56 = errors.New("exit status 56")

// healthyStack is a running stack whose proxy denies and allows correctly.
func healthyStack() *fakeDocker {
	return &fakeDocker{replies: map[string]reply{
		cfgKey:   {out: cfgJSON},
		psKey:    {out: browserUp + "\n" + proxyUp + "\n"},
		denyKey:  {out: "403", err: errExit56},
		allowKey: {out: "200"},
	}}
}

// runStack runs doctor with docker found and d as docker.
func runStack(t *testing.T, d *fakeDocker, vars map[string]string) Report {
	t.Helper()
	root := newWorkspace(t, gateOn, nil)
	opts := options(root, true, &fakeQuarto{pdf: []byte("%PDF-fake")})
	opts.Browser = d
	opts.Lookup = func(k string) (string, bool) {
		if k == "CADDIE_WORKSPACE" {
			return root, true
		}
		v, ok := vars[k]
		return v, ok
	}
	return Run(opts)
}

func wantStatus(t *testing.T, r Report, want map[string]Status) {
	t.Helper()
	for name, s := range want {
		if c := status(r, name); c.Status != s {
			t.Errorf("%s = %+v, want %s", name, c, s)
		}
	}
}

func TestStackHealthy(t *testing.T) {
	d := healthyStack()
	r := runStack(t, d, map[string]string{"CADDIE_COMPOSE_FILE": "/srv/compose.yml"})
	if !r.OK {
		t.Fatalf("expected healthy: %+v", r.Checks)
	}
	wantStatus(t, r, map[string]Status{"docker": StatusOK, "browser-stack": StatusOK,
		"ports": StatusOK, "proxy-deny": StatusOK, "proxy-allow": StatusOK})
	want := []string{"docker", "compose", "-f", "/srv/compose.yml", "exec", "-T", "browser", "curl", "-sS",
		"-o", "/dev/null", "--max-time", "30", "-w", "%{http_connect}", "https://example.com"}
	if len(d.calls) != 4 || !reflect.DeepEqual(d.calls[2], want) {
		t.Errorf("calls = %q\nwant deny probe %q", d.calls, want)
	}
}

func TestStackFailures(t *testing.T) {
	cases := []struct {
		name, key string
		rep       reply
		check     string
		msg       string
	}{
		{"denied host allowed", denyKey, reply{out: "200"}, "proxy-deny", "answered 200"},
		{"listed host denied", allowKey, reply{out: "403", err: errExit56}, "proxy-allow", "want 200"},
		{"proxy unreachable", allowKey, reply{out: "000", err: errors.New("exit status 5")}, "proxy-allow", "no proxy answer"},
		{"published port", psKey, reply{out: browserUp + "\n" +
			`{"Service":"proxy","State":"running","Publishers":[{"TargetPort":8888,"PublishedPort":8888}]}`},
			"ports", "proxy publishes 8888"},
		{"declared port", cfgKey, reply{out: `{"services":{"browser":{"ports":["9222:9222"]},"proxy":{}}}`},
			"ports", "browser declares ports"},
		{"proxy stopped", psKey, reply{out: browserUp}, "browser-stack", "not running: proxy;"},
		{"compose error", cfgKey, reply{err: errors.New("exit status 1")}, "browser-stack", "cannot read"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := healthyStack()
			d.replies[c.key] = c.rep
			r := runStack(t, d, nil)
			got := status(r, c.check)
			if r.OK || got.Status != StatusFail || !strings.Contains(got.Message, c.msg) {
				t.Errorf("%s = %+v, want fail containing %q", c.check, got, c.msg)
			}
		})
	}
}

func TestStackDown(t *testing.T) {
	for _, out := range []string{"", "[]"} {
		d := healthyStack()
		d.replies[psKey] = reply{out: out}
		r := runStack(t, d, nil)
		c := status(r, "browser-stack")
		if r.OK || c.Status != StatusFail || !strings.Contains(c.Message, "up -d --build") {
			t.Errorf("ps %q: browser-stack = %+v", out, c)
		}
		wantStatus(t, r, map[string]Status{"ports": StatusOK})
		if status(r, "proxy-deny").Name != "" || len(d.calls) != 2 {
			t.Errorf("ps %q: probes must not run when the stack is down: %q", out, d.calls)
		}
	}
}

func TestStackPsArray(t *testing.T) {
	d := healthyStack()
	d.replies[psKey] = reply{out: "[" + browserUp + "," + proxyUp + "]"}
	wantStatus(t, runStack(t, d, nil), map[string]Status{"browser-stack": StatusOK})
}

func TestLocalModeSkipsStack(t *testing.T) {
	d := healthyStack()
	r := runStack(t, d, map[string]string{"CADDIE_BROWSER_MODE": "local"})
	if !r.OK || status(r, "browser-stack").Status != StatusWarn || len(d.calls) != 0 {
		t.Errorf("local mode: %+v, calls %q", r.Checks, d.calls)
	}
}

func TestDockerMissingWarns(t *testing.T) {
	d := healthyStack()
	opts := options(newWorkspace(t, gateOn, nil), false, &fakeQuarto{})
	opts.Browser = d
	r := Run(opts)
	if !r.OK || status(r, "docker").Status != StatusWarn || len(d.calls) != 0 {
		t.Errorf("docker missing: %+v, calls %q", r.Checks, d.calls)
	}
}
