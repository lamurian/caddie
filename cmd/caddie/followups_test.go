package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fuWorkspace is the fixture workspace (deadline_warning_days: 5; the test
// clock says 2026-10-28) with the given files written under its root.
func fuWorkspace(t *testing.T, files map[string]string) map[string]string {
	t.Helper()
	prefs, err := os.ReadFile("../../testdata/config/preferences.fixture.yml")
	if err != nil {
		t.Fatal(err)
	}
	vars := workspaceWith(t, string(prefs), nil)
	for rel, content := range files {
		writeFile(t, filepath.Join(vars["CADDIE_WORKSPACE"], rel), content)
	}
	return vars
}

// fuPosting is a minimal posting; extra holds further frontmatter lines.
func fuPosting(slug, status, extra, body string) string {
	return "---\nslug: " + slug + "\ntitle: Analyst\ncompany: Initech\n" +
		"url: https://example.com/" + slug + "\nstatus: " + status + "\n" + extra + "---\n" + body
}

func fuAll(t *testing.T) map[string]string {
	t.Helper()
	return fuWorkspace(t, map[string]string{
		"applied/due.md": fuPosting("due", "applied",
			"applied_at: 2026-10-07\nfollow_up_after: 2026-10-21\n", ""),
		"applied/later.md": fuPosting("later", "applied",
			"applied_at: 2026-10-27\nfollow_up_after: 2026-11-10\n", ""),
		"manual/steps.md": fuPosting("steps", "manual", "manual_reason: login wall\n",
			"## Manual steps\n\n- [x] Create account\n- [ ] Upload the resume\n- [ ] Send email\n"),
		"postings/gone.md": fuPosting("gone", "saved", "deadline: 2026-10-27\n", ""),
		"postings/soon.md": fuPosting("soon", "saved", "deadline: 2026-10-30\n", ""),
		"postings/far.md":  fuPosting("far", "saved", "deadline: 2026-12-01\n", ""),
	})
}

const fuWantJSON = `{
  "today": "2026-10-28",
  "deadline_warning_days": 5,
  "follow_ups": [
    {
      "slug": "due",
      "title": "Analyst",
      "company": "Initech",
      "posting": "applied/due.md",
      "applied_at": "2026-10-07",
      "follow_up_after": "2026-10-21",
      "days_overdue": 7
    }
  ],
  "manual_steps": [
    {
      "slug": "steps",
      "title": "Analyst",
      "company": "Initech",
      "posting": "manual/steps.md",
      "manual_reason": "login wall",
      "open": 2,
      "items": [
        "Upload the resume",
        "Send email"
      ]
    }
  ],
  "deadlines": [
    {
      "slug": "gone",
      "title": "Analyst",
      "company": "Initech",
      "posting": "postings/gone.md",
      "status": "saved",
      "deadline": "2026-10-27",
      "days_left": -1,
      "past": true
    },
    {
      "slug": "soon",
      "title": "Analyst",
      "company": "Initech",
      "posting": "postings/soon.md",
      "status": "saved",
      "deadline": "2026-10-30",
      "days_left": 2,
      "past": false
    }
  ]
}
`

func TestFollowUpsJSON(t *testing.T) {
	r := runCaddie(t, fuAll(t), "follow-ups", "--json")
	if r.code != exitOK || r.stderr != "" {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	if r.stdout != fuWantJSON {
		t.Errorf("got:\n%s\nwant:\n%s", r.stdout, fuWantJSON)
	}
}

func TestFollowUpsText(t *testing.T) {
	r := runCaddie(t, fuAll(t), "follow-ups")
	want := `Follow-ups due (2026-10-28):
  applied/due.md  Analyst at Initech  applied 2026-10-07, follow up after 2026-10-21 (7 days overdue)
Open manual checklists:
  manual/steps.md  Analyst at Initech  2 open (login wall)
    - [ ] Upload the resume
    - [ ] Send email
Deadlines (within 5 days):
  postings/gone.md  Analyst at Initech  deadline 2026-10-27  PASSED 1 day ago
  postings/soon.md  Analyst at Initech  deadline 2026-10-30  2 days left
`
	if r.code != exitOK || r.stdout != want {
		t.Errorf("exit %d: got:\n%s\nwant:\n%s%s", r.code, r.stdout, want, r.stderr)
	}
}

func TestFollowUpsEmpty(t *testing.T) {
	vars := fuWorkspace(t, nil)
	if r := runCaddie(t, vars, "follow-ups"); r.code != exitOK || r.stdout != "Nothing to follow up.\n" {
		t.Errorf("text: exit %d: %q %s", r.code, r.stdout, r.stderr)
	}
	r := runCaddie(t, vars, "follow-ups", "--json")
	if r.code != exitOK || !strings.Contains(r.stdout, `"follow_ups": [],`) ||
		!strings.Contains(r.stdout, `"manual_steps": [],`) || !strings.Contains(r.stdout, `"deadlines": []`) {
		t.Errorf("json: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
}

func TestFollowUpsErrors(t *testing.T) {
	bad := fuWorkspace(t, map[string]string{
		"postings/bad.md": fuPosting("bad", "saved", "deadline: next week\n", ""),
	})
	r := runCaddie(t, bad, "follow-ups")
	if r.code != exitMalformed || !strings.Contains(r.stderr, "postings/bad.md") || r.stdout != "" {
		t.Errorf("malformed: exit %d: %q %s", r.code, r.stdout, r.stderr)
	}
	vars := fuWorkspace(t, nil)
	if r := runCaddie(t, vars, "follow-ups", "--bogus"); r.code != exitUsage {
		t.Errorf("bad flag: exit %d: %s", r.code, r.stderr)
	}
	if r := runCaddie(t, vars, "follow-ups", "extra"); r.code != exitUsage {
		t.Errorf("extra arg: exit %d: %s", r.code, r.stderr)
	}
	root := vars["CADDIE_WORKSPACE"]
	writeFile(t, filepath.Join(root, "preferences.yml"), "apply: [not, a, map]\n")
	if r := runCaddie(t, vars, "follow-ups"); r.code != exitConfig {
		t.Errorf("bad preferences: exit %d: %s", r.code, r.stderr)
	}
}
