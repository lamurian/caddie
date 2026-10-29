package document

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/lamurian/caddie/internal/profile"
)

// verifyTargets are the company and title of the fictional posting.
var verifyTargets = []string{"Example Health Alliance", "Senior Biostatistician"}

// letterAnswers are the answers cover-letter-answers.md draws on.
var letterAnswers = []string{
	"The Example Health Alliance's open-data mandate matches how I already work.",
	"In 2021 I advised the Bergen Outbreak Taskforce on tuberculosis cohort data.",
	"2027-01-01",
}

func testProfile(t *testing.T) profile.Profile {
	t.Helper()
	p, err := profile.Load(filepath.Join("..", "..", "testdata", "history"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func readDraft(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "documents", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Each fixture fails on exactly the claims it invents, with line and kind.
func TestVerifyFixtures(t *testing.T) {
	p := testProfile(t)
	cases := []struct {
		file    string
		answers []string
		want    []Claim
	}{
		{"resume-traceable.md", nil, []Claim{}},
		{"resume-invented-employer.md", nil, []Claim{
			{6, KindEntry, "Acme Pharma Analytics"},
			{12, KindTerm, "Fjordline Clinical Institute"},
		}},
		{"resume-stretched-dates.md", nil, []Claim{
			{7, KindEntry, "Sep 2018 – Feb 2023"},
			{8, KindDate, "June 2018"},
			{15, KindEntry, "Sep 2017 – Jun 2019"},
			{16, KindDate, "September 2017"},
		}},
		{"resume-upgraded-title.md", nil, []Claim{
			{5, KindEntry, "Senior Data Analyst"},
			{11, KindEntry, "Fjordline Health Research"},
			{12, KindEntry, "Utrecht, Netherlands"},
			{13, KindEntry, "Mar 2023 – Present"},
		}},
		{"resume-invented-skill.md", nil, []Claim{
			{3, KindTerm, "PyTorch"},
			{4, KindTerm, "Kubernetes"},
			{8, KindSkill, "deep learning"},
			{9, KindSkill, "Kubernetes"},
		}},
		{"cover-letter-answers.md", letterAnswers, []Claim{}},
		{"cover-letter-answers.md", nil, []Claim{
			{9, KindTerm, "Bergen Outbreak Taskforce"},
			{14, KindDate, "1 January 2027"},
		}},
	}
	for _, c := range cases {
		src := Sources{Profile: p, Answers: c.answers, Targets: verifyTargets}
		r := Verify(readDraft(t, c.file), src)
		if !reflect.DeepEqual(r.Untraceable, c.want) {
			t.Errorf("%s (answers: %d): untraceable\n got %+v\nwant %+v", c.file, len(c.answers), r.Untraceable, c.want)
		}
		if r.OK() != (len(c.want) == 0) || r.Claims < len(c.want) || r.Claims == 0 {
			t.Errorf("%s: OK %v with %d claims", c.file, r.OK(), r.Claims)
		}
	}
}

// The traceable resume exercises every claim kind, so a pass means each
// kind was extracted and matched, not skipped.
func TestVerifyCountsEveryKind(t *testing.T) {
	r := Verify(readDraft(t, "resume-traceable.md"), Sources{Profile: testProfile(t)})
	if !r.OK() || r.Claims < 30 {
		t.Errorf("claims %d, untraceable %+v", r.Claims, r.Untraceable)
	}
}
