package doctor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lamurian/caddie/internal/browse"
	"github.com/lamurian/caddie/internal/config"
)

// dockerTimeout bounds each docker call; probeTimeout is curl's own limit
// inside the browser container, kept below it.
const (
	dockerTimeout = 45 * time.Second
	probeTimeout  = "30"
)

// The proxy probes (ADR 0023): a host the allowlist must deny, and one it
// must allow. The verdict is curl's %{http_connect}, the proxy's answer to
// CONNECT, so TLS or the site itself cannot sway it.
const (
	deniedURL   = "https://example.com"
	allowedURL  = "https://careers.who.int/"
	deniedCode  = "403"
	allowedCode = "200"
)

// stack runs docker compose against one compose file.
type stack struct {
	ex      browse.Exec
	file    string
	service string
}

// checkBrowser checks the sandboxed browser stack (ADR 0002, ADR 0023).
func checkBrowser(r *Report, opts Options, cfg config.Config) {
	if cfg.BrowserMode != "compose" {
		r.add("browser-stack", StatusWarn, "CADDIE_BROWSER_MODE=%s: agent-browser runs outside "+
			"the sandbox; stack checks skipped", cfg.BrowserMode)
		return
	}
	path, err := opts.LookPath("docker")
	if err != nil {
		r.add("docker", StatusWarn, "docker not found on PATH; needed by `caddie browse`; "+
			"stack checks skipped")
		return
	}
	r.add("docker", StatusOK, "found at %s", path)
	s := stack{ex: opts.Browser, file: cfg.Resolve(cfg.ComposeFile), service: cfg.ComposeService}
	if s.ex == nil {
		s.ex = browse.OSExec{}
	}
	if s.checkUp(r) {
		s.checkProbe(r, "proxy-deny", deniedURL, deniedCode)
		s.checkProbe(r, "proxy-allow", allowedURL, allowedCode)
	}
}

// composeConfig is the part of `docker compose config --format json` read.
type composeConfig struct {
	Services map[string]struct {
		Ports []json.RawMessage `json:"ports"`
	} `json:"services"`
}

// container is the part of one `docker compose ps --format json` entry read.
type container struct {
	Service    string `json:"Service"`
	State      string `json:"State"`
	Publishers []struct {
		PublishedPort int `json:"PublishedPort"`
	} `json:"Publishers"`
}

// checkUp adds `browser-stack` and `ports` and reports whether every
// service is running, which the proxy probes need.
func (s stack) checkUp(r *Report) bool {
	var cfg composeConfig
	out, err := s.run("config", "--format", "json")
	if err == nil {
		err = json.Unmarshal(out, &cfg)
	}
	if err != nil {
		r.add("browser-stack", StatusFail, "cannot read %s: %v", s.file, err)
		return false
	}
	out, err = s.run("ps", "--format", "json")
	var cs []container
	if err == nil {
		cs, err = parseContainers(out)
	}
	if err != nil {
		r.add("browser-stack", StatusFail, "docker compose ps failed: %v", err)
		return false
	}
	up := s.reportRunning(r, cfg, cs)
	reportPorts(r, cfg, cs)
	return up
}

// reportRunning adds `browser-stack`: every service must be running.
func (s stack) reportRunning(r *Report, cfg composeConfig, cs []container) bool {
	running := map[string]bool{}
	for _, c := range cs {
		if c.State == "running" {
			running[c.Service] = true
		}
	}
	var all, down []string
	for name := range cfg.Services {
		all = append(all, name)
		if !running[name] {
			down = append(down, name)
		}
	}
	sort.Strings(all)
	sort.Strings(down)
	if len(down) > 0 {
		r.add("browser-stack", StatusFail, "not running: %s; start the stack: "+
			"docker compose -f %s up -d --build", strings.Join(down, ", "), s.file)
		return false
	}
	r.add("browser-stack", StatusOK, "running: %s", strings.Join(all, ", "))
	return true
}

// reportPorts adds `ports`: no service may declare or publish a host port.
func reportPorts(r *Report, cfg composeConfig, cs []container) {
	var bad []string
	for name, svc := range cfg.Services {
		if len(svc.Ports) > 0 {
			bad = append(bad, name+" declares ports")
		}
	}
	for _, c := range cs {
		for _, p := range c.Publishers {
			if p.PublishedPort != 0 {
				bad = append(bad, fmt.Sprintf("%s publishes %d", c.Service, p.PublishedPort))
			}
		}
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		r.add("ports", StatusFail, "%s; remove every `ports:` (PRD §10)", strings.Join(bad, "; "))
		return
	}
	r.add("ports", StatusOK, "no service publishes a port")
}

// parseContainers reads `ps --format json`: one object per line (Compose
// v2.21+) or a single array (older releases). No output means none.
func parseContainers(out []byte) ([]container, error) {
	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil, nil
	}
	var cs []container
	if out[0] == '[' {
		return cs, json.Unmarshal(out, &cs)
	}
	dec := json.NewDecoder(bytes.NewReader(out))
	for dec.More() {
		var c container
		if err := dec.Decode(&c); err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, nil
}

// checkProbe curls url from the browser container through the proxy and
// fails unless the proxy answers CONNECT with want.
func (s stack) checkProbe(r *Report, name, url, want string) {
	out, err := s.run("exec", "-T", s.service, "curl", "-sS", "-o", "/dev/null",
		"--max-time", probeTimeout, "-w", "%{http_connect}", url)
	got := strings.TrimSpace(string(out))
	switch {
	case got == want:
		r.add(name, StatusOK, "proxy answered %s for %s", got, url)
	case len(got) == 3 && got != "000":
		r.add(name, StatusFail, "proxy answered %s for %s, want %s; "+
			"check docker/proxy/allowlist.conf", got, url, want)
	default:
		r.add(name, StatusFail, "probe of %s got no proxy answer: %v", url, err)
	}
}

// run runs `docker compose -f <file> <args…>` and returns its stdout; an
// error carries the tail of stderr.
func (s stack) run(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dockerTimeout)
	defer cancel()
	argv := append([]string{"docker", "compose", "-f", s.file}, args...)
	var stdout, stderr bytes.Buffer
	if err := s.ex.Run(ctx, argv, &stdout, &stderr); err != nil {
		return stdout.Bytes(), fmt.Errorf("%w%s", err, outputTail(stderr.Bytes()))
	}
	return stdout.Bytes(), nil
}
