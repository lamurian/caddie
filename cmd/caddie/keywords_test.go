package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/document"
)

const (
	keywordsPosting = "../../testdata/postings/lead-biostatistician-at-globex.md"
	keywordsDraft   = "../../testdata/documents/keywords-draft-lead-biostatistician.md"
	// Owned (and rewritten under -update) by internal/document's golden test.
	keywordsGolden = "../../testdata/golden/keywords-lead-biostatistician-at-globex.txt"
)

func TestKeywordsCommand(t *testing.T) {
	want, err := os.ReadFile(keywordsGolden)
	if err != nil {
		t.Fatal(err)
	}
	r := runCaddie(t, fixtureVars, "keywords", keywordsPosting, keywordsDraft)
	if r.code != exitOK || r.stderr != "" || r.stdout != string(want) {
		t.Errorf("exit %d, output differs from %s:\n%s%s", r.code, keywordsGolden, r.stdout, r.stderr)
	}
	r = runCaddie(t, fixtureVars, "keywords", keywordsPosting, "--json", keywordsDraft)
	var rep document.KeywordReport
	if r.code != exitOK || json.Unmarshal([]byte(r.stdout), &rep) != nil {
		t.Fatalf("--json: exit %d: %s%s", r.code, r.stdout, r.stderr)
	}
	if rep.Total == 0 || rep.Total != len(rep.Keywords) || rep.Text() != string(want) {
		t.Errorf("--json does not round-trip to the text report: %+v", rep)
	}
}

func TestKeywordsErrors(t *testing.T) {
	cases := []struct {
		args []string
		code int
		want string
	}{
		{[]string{"keywords"}, exitUsage, "keywords <posting.md> <doc.md>"},
		{[]string{"keywords", keywordsPosting}, exitUsage, "keywords <posting.md> <doc.md>"},
		{[]string{"keywords", "--bogus", keywordsPosting, keywordsDraft}, exitUsage, "flag provided but not defined"},
		{[]string{"keywords", "absent.md", keywordsDraft}, exitNotFound, `no file "absent.md"`},
		{[]string{"keywords", keywordsPosting, "absent.md"}, exitNotFound, `no file "absent.md"`},
	}
	for _, c := range cases {
		r := runCaddie(t, fixtureVars, c.args...)
		if r.code != c.code || !strings.Contains(r.stderr, c.want) {
			t.Errorf("%v: exit %d: %s", c.args, r.code, r.stderr)
		}
	}
}
