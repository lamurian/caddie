package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files in testdata/golden")

const golden = "../../testdata/golden/profile.json"

// fixtureVars point the CLI at the fictional workspace in testdata/.
var fixtureVars = map[string]string{
	"CADDIE_WORKSPACE":   "../../testdata",
	"CADDIE_PREFERENCES": "config/preferences.fixture.yml",
}

func TestProfileGolden(t *testing.T) {
	r := runCaddie(t, fixtureVars, "profile", "--json")
	if r.code != exitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	if *update {
		if err := os.WriteFile(golden, []byte(r.stdout), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if r.stdout != string(want) {
		t.Errorf("caddie profile --json differs from %s; rerun with -update and review the diff\n%s",
			golden, r.stdout)
	}
	if plain := runCaddie(t, fixtureVars, "profile"); plain.stdout != r.stdout {
		t.Error("profile output must not depend on --json")
	}
}

func TestProfileOmitsContactByDefault(t *testing.T) {
	r := runCaddie(t, fixtureVars, "profile")
	for _, secret := range []string{"mira@example.org", "+00 0 00000000", `"contact"`, "email", "phone"} {
		if strings.Contains(r.stdout, secret) {
			t.Errorf("default profile output contains %q", secret)
		}
	}
	r = runCaddie(t, fixtureVars, "profile", "--with-contact")
	if !strings.Contains(r.stdout, `"email": "mira@example.org"`) || !strings.Contains(r.stdout, "+00 0 00000000") {
		t.Errorf("--with-contact output lacks contact details:\n%s", r.stdout)
	}
}

func TestProfileErrors(t *testing.T) {
	root := t.TempDir()
	hist := filepath.Join(root, "history")
	if err := os.Mkdir(hist, 0o755); err != nil {
		t.Fatal(err)
	}
	vars := map[string]string{"CADDIE_WORKSPACE": root}
	if r := runCaddie(t, vars, "profile"); r.code != exitMalformed || !strings.Contains(r.stderr, "about.md: missing") {
		t.Errorf("missing about.md: %+v", r)
	}
	about := "---\nname: Test Person\n---\n## About\nFictional.\n"
	if err := os.WriteFile(filepath.Join(hist, "about.md"), []byte(about), 0o600); err != nil {
		t.Fatal(err)
	}
	if r := runCaddie(t, vars, "profile"); r.code != exitConfig || !strings.Contains(r.stderr, "preferences.example.yml") {
		t.Errorf("missing preferences: %+v", r)
	}
}
