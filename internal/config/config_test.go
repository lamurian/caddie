package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func lookupFrom(m map[string]string) LookupFunc {
	return func(k string) (string, bool) { v, ok := m[k]; return v, ok }
}

func TestLoadPrecedence(t *testing.T) {
	dotenv := writeFile(t, "CADDIE_WORKSPACE=from-dotenv\n"+
		"CADDIE_STATE_DIR=dotenv-state\n"+
		"CADDIE_QUARTO_BIN=  # empty value falls back to default\n"+
		"OTHER_TOOL_KEY=ignored\n")
	env := map[string]string{"CADDIE_WORKSPACE": "from-env", "CADDIE_LOG_LEVEL": "debug"}
	cfg, err := Load(dotenv, lookupFrom(env))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, got, want string }{
		{"env beats .env", cfg.Workspace, "from-env"},
		{"env beats default", cfg.LogLevel, "debug"},
		{".env beats default", cfg.StateDir, "dotenv-state"},
		{"empty falls back to default", cfg.QuartoBin, "quarto"},
		{"default when unset", cfg.ComposeService, "browser"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
	if cfg.BrowseTimeout != 60*time.Second {
		t.Errorf("BrowseTimeout = %v", cfg.BrowseTimeout)
	}
}

func TestLoadMissingDotEnvUsesDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "absent.env"), lookupFrom(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Workspace != "." || cfg.Preferences != "preferences.yml" || cfg.HandoffDir != "handoff" {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
}

func TestDefaultsMatchEnvExample(t *testing.T) {
	example, err := ReadDotEnv("../../.env.example")
	if err != nil {
		t.Fatal(err)
	}
	if len(example) != len(Defaults) {
		t.Errorf(".env.example has %d keys, Defaults has %d", len(example), len(Defaults))
	}
	for key, want := range Defaults {
		if got, ok := example[key]; !ok || got != want {
			t.Errorf("%s: .env.example has %q, default is %q", key, got, want)
		}
	}
}

func TestDotEnvSyntax(t *testing.T) {
	got, err := ParseDotEnv("x.env", []byte("# comment\n\nexport A=1\nB=\"two # not a comment\"\n"+
		"C='three'  # comment\nD=four\t# tab comment\nE=\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"A": "1", "B": "two # not a comment", "C": "three", "D": "four", "E": ""}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
}

func TestDotEnvErrorsNameFileAndLine(t *testing.T) {
	cases := map[string]string{
		"A=1\nnot a pair\n":  "x.env:2:",
		"A=1\nB=2\n9X=3\n":   "x.env:3:",
		"A=\"unterminated\n": "x.env:1:",
		"A='q' trailing\n":   "x.env:1:",
	}
	for input, want := range cases {
		_, err := ParseDotEnv("x.env", []byte(input))
		if err == nil || !strings.Contains(err.Error(), want) || !errors.Is(err, ErrConfig) {
			t.Errorf("%q: got %v, want error containing %q", input, err, want)
		}
	}
}

func TestLoadRejectsBadValues(t *testing.T) {
	cases := map[string]map[string]string{
		"unknown key":   nil,
		"browser mode":  {"CADDIE_BROWSER_MODE": "remote"},
		"timeout":       {"CADDIE_BROWSE_TIMEOUT": "soon"},
		"zero timeout":  {"CADDIE_BROWSE_TIMEOUT": "0s"},
		"log format":    {"CADDIE_LOG_FORMAT": "xml"},
		"log level bad": {"CADDIE_LOG_LEVEL": "loud"},
	}
	for name, env := range cases {
		dotenv := filepath.Join(t.TempDir(), "absent.env")
		if env == nil {
			dotenv = writeFile(t, "CADDIE_WORKSPCE=typo\n")
		}
		if _, err := Load(dotenv, lookupFrom(env)); !errors.Is(err, ErrConfig) {
			t.Errorf("%s: got %v, want ErrConfig", name, err)
		}
	}
}

func TestResolve(t *testing.T) {
	c := Config{Workspace: "ws"}
	if got := c.Resolve("handoff"); got != filepath.Join("ws", "handoff") {
		t.Errorf("relative: %q", got)
	}
	if got := c.Resolve("/abs/x"); got != "/abs/x" {
		t.Errorf("absolute: %q", got)
	}
}
