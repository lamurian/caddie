package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/document"
	"github.com/lamurian/caddie/internal/posting"
	"github.com/lamurian/caddie/internal/profile"
	"github.com/lamurian/caddie/internal/source"
	"github.com/lamurian/caddie/internal/workspace"
)

type result struct {
	code           int
	stdout, stderr string
}

// runCaddie runs the CLI in-process with only the given env vars set, no
// .env file and no network.
func runCaddie(t *testing.T, vars map[string]string, args ...string) result {
	t.Helper()
	return runWith(t, vars, fakeHTTP(nil), args...)
}

// runWith is runCaddie with HTTP answered from recorded responses.
func runWith(t *testing.T, vars map[string]string, client fakeHTTP, args ...string) result {
	t.Helper()
	return runQuarto(t, vars, client, nil, args...)
}

// runQuarto is runWith with q as quarto; a nil q fails the test if run.
func runQuarto(t *testing.T, vars map[string]string, client fakeHTTP, q document.Quarto, args ...string) result {
	t.Helper()
	if q == nil {
		q = noQuarto{t}
	}
	var out, errOut bytes.Buffer
	code := run(args, env{
		stdout: &out, stderr: &errOut,
		lookup: func(k string) (string, bool) { v, ok := vars[k]; return v, ok },
		dotenv: filepath.Join(t.TempDir(), "absent.env"),
		lookPath: func(f string) (string, error) {
			return "", errors.New("not found")
		},
		client: client,
		today:  func() string { return "2026-10-28" },
		quarto: func(string) document.Quarto { return q },
	})
	return result{code, out.String(), errOut.String()}
}

// fakeHTTP maps a URL to a file in testdata/sources/; other URLs get 404.
type fakeHTTP map[string]string

func (f fakeHTTP) Do(req *http.Request) (*http.Response, error) {
	file, ok := f[req.URL.String()]
	if !ok {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil
	}
	data, err := os.ReadFile(filepath.Join("../../testdata/sources", file))
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(data))}, err
}

func TestUsage(t *testing.T) {
	cases := []struct {
		args []string
		code int
		want string
	}{
		{nil, exitUsage, "Usage: caddie"},
		{[]string{"help"}, exitOK, "Commands:"},
		{[]string{"frobnicate"}, exitUsage, `unknown command "frobnicate"`},
		{[]string{"profile", "--bogus"}, exitUsage, "flag provided but not defined"},
		{[]string{"doctor", "extra"}, exitUsage, `unexpected argument "extra"`},
		{[]string{"doctor", "-h"}, exitOK, "-json"},
		{[]string{"search", "--source", "bogus"}, exitUsage, `unknown source "bogus"`},
		{[]string{"search", "--source", "portal:WHO"}, exitUsage, "portal[:<org>]"},
		{[]string{"search", "--limit", "-1"}, exitUsage, "--limit"},
		{[]string{"seen"}, exitUsage, "seen add <id…>"},
		{[]string{"seen", "forget"}, exitUsage, `unknown subcommand "forget"`},
		{[]string{"seen", "add", "a:1"}, exitUsage, "exactly one of --shown or --rejected"},
		{[]string{"seen", "add", "a:1", "--shown", "--rejected"}, exitUsage, "exactly one of"},
		{[]string{"seen", "add", "--shown"}, exitUsage, "at least one job ID"},
		{[]string{"seen", "list", "extra"}, exitUsage, `unexpected argument "extra"`},
		{[]string{"posting"}, exitUsage, "posting save <url>"},
		{[]string{"posting", "save"}, exitUsage, "exactly one URL"},
		{[]string{"posting", "save", "postings/x.md"}, exitUsage, "not an http(s) URL"},
	}
	for _, c := range cases {
		r := runCaddie(t, nil, c.args...)
		if r.code != c.code || !strings.Contains(r.stdout+r.stderr, c.want) {
			t.Errorf("%v: code %d, output %q; want %d and %q", c.args, r.code, r.stdout+r.stderr, c.code, c.want)
		}
	}
}

func TestExitCodeMapping(t *testing.T) {
	cases := []struct {
		err  error
		code int
	}{
		{nil, exitOK},
		{errCheckFailed, exitCheckFailed},
		{usageErr("x"), exitUsage},
		{config.ErrConfig, exitConfig},
		{&profile.ParseError{File: "f.md", Line: 1, Msg: "x"}, exitMalformed},
		{workspace.ErrBadPosting, exitMalformed},
		{workspace.Transition(workspace.StatusSaved, workspace.StatusApplied), exitIllegalState},
		{fmt.Errorf("x: %w", posting.ErrNoPosting), exitNotFound},
		{fmt.Errorf("x: %w", source.ErrFetch), exitExternal},
		{fmt.Errorf("x: %w", document.ErrNotVerified), exitRefused},
		{fmt.Errorf("x: %w", document.ErrQuarto), exitExternal},
		{fmt.Errorf("x: %w", document.ErrKind), exitUsage},
		{errors.New("surprise"), exitInternal},
	}
	for _, c := range cases {
		if got := exitCode(c.err); got != c.code {
			t.Errorf("exitCode(%v) = %d, want %d", c.err, got, c.code)
		}
	}
}

func TestConfigErrorExitCode(t *testing.T) {
	r := runCaddie(t, map[string]string{"CADDIE_BROWSER_MODE": "remote"}, "profile")
	if r.code != exitConfig || !strings.Contains(r.stderr, "CADDIE_BROWSER_MODE") {
		t.Errorf("got %+v", r)
	}
}
