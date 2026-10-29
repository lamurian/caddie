package document

import (
	"bytes"
	"flag"
	"os"
	"testing"

	"github.com/lamurian/caddie/internal/profile"
)

// update rewrites every golden file this package owns in testdata/golden.
var update = flag.Bool("update", false, "rewrite the golden reports in testdata/golden")

// checkGolden compares got with the golden file at path, rewriting it
// first under -update.
func checkGolden(t *testing.T, path string, got []byte) {
	t.Helper()
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("output differs from %s; rerun with -update and review the diff\n%s", path, got)
	}
}

// fixtureProfile loads the fictional testdata/history.
func fixtureProfile(t *testing.T) profile.Profile {
	t.Helper()
	p, err := profile.Load("../../testdata/history")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestKeywordsGolden(t *testing.T) {
	posting := readFixture(t, "../../testdata/postings/lead-biostatistician-at-globex.md")
	draft := readFixture(t, "../../testdata/documents/keywords-draft-lead-biostatistician.md")
	r := Keywords(posting, draft, fixtureProfile(t))
	checkGolden(t, "../../testdata/golden/keywords-lead-biostatistician-at-globex.txt", []byte(r.Text()))
}
