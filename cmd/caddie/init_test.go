package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
)

// runInit runs caddie in an empty workspace root with stdin; .env and
// .env.example live in root too.
func runInit(t *testing.T, root, stdin string, args ...string) result {
	t.Helper()
	var out, errOut bytes.Buffer
	code := run(args, env{
		stdin: strings.NewReader(stdin), stdout: &out, stderr: &errOut,
		lookup: func(k string) (string, bool) { return root, k == "CADDIE_WORKSPACE" },
		dotenv: filepath.Join(root, ".env"),
	})
	return result{code, out.String(), errOut.String()}
}

func initRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	copyFile(t, "../../.env.example", filepath.Join(root, ".env.example"))
	for _, p := range []string{"who", "un"} {
		copyFile(t, "../../portals/"+p+".yml", filepath.Join(root, "portals", p+".yml"))
	}
	return root
}

func TestInitInteractive(t *testing.T) {
	root := initRoot(t)
	r := runInit(t, root, "me@example.org\n\ndata scientist\n", "init")
	if r.code != exitOK {
		t.Fatalf("exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	for _, want := range []string{"created " + filepath.Join(root, ".env"), "ok   preferences: preferences.yml",
		"ok   answers: answers.yml", "Organisation portals to search (un, who) [un, who]"} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("stdout lacks %q:\n%s", want, r.stdout)
		}
	}
	p, err := config.LoadPreferences(filepath.Join(root, "preferences.yml"))
	if err != nil || p.Contact.Email != "me@example.org" || p.Roles.Include[0] != "data scientist" {
		t.Errorf("preferences: %+v, %v", p, err)
	}
	if r := runInit(t, root, "", "validate"); r.code != exitOK || strings.Count(r.stdout, "ok ") != 2 {
		t.Errorf("validate: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	// A second run changes nothing and asks nothing.
	if r := runInit(t, root, "", "init"); r.code != exitOK || strings.Contains(r.stdout, "Answer each") {
		t.Errorf("rerun: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
}

func TestInitQuestionsAndAnswers(t *testing.T) {
	root := initRoot(t)
	r := runInit(t, root, "", "init", "--questions")
	var q questionsOutput
	if err := json.Unmarshal([]byte(r.stdout), &q); err != nil || r.code != exitOK || len(q.Missing) != 2 || len(q.Questions) < 20 {
		t.Fatalf("questions: exit %d, %v: %s", r.code, err, r.stdout)
	}
	if _, err := os.Stat(filepath.Join(root, ".env")); err == nil {
		t.Error("--questions wrote .env")
	}
	r = runInit(t, root, `{"roles.include": ["epidemiologist"], "role.include": "x"}`, "init", "--answers", "-")
	if r.code != exitUsage || !strings.Contains(r.stderr, `unknown question id "role.include"`) {
		t.Errorf("typo: exit %d: %s", r.code, r.stderr)
	}
	r = runInit(t, root, `{"roles.include": ["epidemiologist"], "global.notice_period": "1 month"}`, "init", "--answers", "-")
	if r.code != exitOK {
		t.Fatalf("answers: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	data, _ := os.ReadFile(filepath.Join(root, "answers.yml"))
	if !strings.Contains(string(data), "notice_period: 1 month") {
		t.Errorf("answers.yml:\n%s", data)
	}
	r = runInit(t, root, "", "init", "--questions")
	if !strings.Contains(r.stdout, `"missing": []`) || !strings.Contains(r.stdout, `"questions": []`) {
		t.Errorf("nothing missing: %s", r.stdout)
	}
}

func TestInitRequiredAtEOF(t *testing.T) {
	r := runInit(t, initRoot(t), "", "init")
	if r.code != exitUsage || !strings.Contains(r.stderr, "roles.include: an answer is required") {
		t.Errorf("exit %d: %s", r.code, r.stderr)
	}
}

func TestValidate(t *testing.T) {
	root := initRoot(t)
	r := runInit(t, root, "", "validate")
	if r.code != exitConfig || !strings.Contains(r.stdout, "fail preferences") || !strings.Contains(r.stderr, "run caddie init") {
		t.Errorf("missing: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("answers.yml", "global: [x]\n")
	if r := runInit(t, root, "", "validate", "answers"); r.code != exitMalformed {
		t.Errorf("malformed answers: exit %d: %s", r.code, r.stderr)
	}
	for content, want := range map[string]string{
		"apply:\n  require_gat: true\n":   "require_gat",
		"apply:\n  require_gate: false\n": "must be true",
		"work_mode: [office]\n":           "work_mode",
		"salary:\n  minimum: -1\n":        "must not be negative",
	} {
		write("preferences.yml", content)
		if r := runInit(t, root, "", "validate", "preferences"); r.code != exitConfig || !strings.Contains(r.stderr, want) {
			t.Errorf("%q: exit %d: %s", content, r.code, r.stderr)
		}
	}
	if r := runInit(t, root, "", "validate", "history"); r.code != exitUsage {
		t.Errorf("unknown file: exit %d", r.code)
	}
}
