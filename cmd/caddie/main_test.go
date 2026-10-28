package main

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/profile"
	"github.com/lamurian/caddie/internal/workspace"
)

type result struct {
	code           int
	stdout, stderr string
}

// runCaddie runs the CLI in-process with only the given env vars set and
// no .env file.
func runCaddie(t *testing.T, vars map[string]string, args ...string) result {
	t.Helper()
	var out, errOut bytes.Buffer
	code := run(args, env{
		stdout: &out, stderr: &errOut,
		lookup: func(k string) (string, bool) { v, ok := vars[k]; return v, ok },
		dotenv: filepath.Join(t.TempDir(), "absent.env"),
		lookPath: func(f string) (string, error) {
			return "", errors.New("not found")
		},
	})
	return result{code, out.String(), errOut.String()}
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
