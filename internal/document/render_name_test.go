package document

import (
	"errors"
	"strings"
	"testing"
)

func TestPDFName(t *testing.T) {
	longName := strings.Repeat("Bartholomew ", 10)
	cappedName := strings.TrimSuffix(strings.Repeat("Bartholomew-", 5), "-")
	longPos := strings.Repeat("principal ", 20)
	cappedPos := strings.TrimSuffix(strings.Repeat("principal-", 8), "-")
	cases := []struct {
		name, position, company string
		want                    string // the part after <Name>-<kind>-
		wantName                string
	}{
		{"Mira Example-Halvorsen", "Senior Data Scientist", "Acme", "senior-data-scientist-acme", "Mira-Example-Halvorsen"},
		{"Mira", "Analyst", "Globex", "analyst-globex", "Mira"},
		{"José Ñúñez O'Brien", "Data Scientist", "Acme", "data-scientist-acme", "Jose-Nunez-O-Brien"},
		{"John Smith, Jr.", "Statistician", "Acme", "statistician-acme", "John-Smith-Jr"},
		{"  Anne-Marie   d’Arcy  ", "Researcher", "Acme", "researcher-acme", "Anne-Marie-d-Arcy"},
		{"Zoë Ångström", "Épidémiologiste", "Société Générale", "epidemiologiste-societe-generale", "Zoe-Angstrom"},
		{"Mira Example", "R&D / ML Engineer (Remote)", "Procter & Gamble", "r-d-ml-engineer-remote-procter-gamble", "Mira-Example"},
		{"Mira Example", "Data Scientist", "Ernst & Young (EY) Nederland", "data-scientist-ernst-young-ey-nederland", "Mira-Example"},
		{"Mira Example", "Head of Data/AI", "AT&T", "head-of-data-ai-at-t", "Mira-Example"},
		{longName, longPos, "Acme", cappedPos + "-acme", cappedName},
	}
	for _, c := range cases {
		for _, kind := range []Kind{KindResume, KindCoverLetter} {
			want := c.wantName + "-" + string(kind) + "-" + c.want + ".pdf"
			if got := PDFName(c.name, kind, c.position, c.company); got != want {
				t.Errorf("PDFName(%q, %s, %q, %q) = %q, want %q", c.name, kind, c.position, c.company, got, want)
			}
		}
	}
}

func TestPDFNameEdges(t *testing.T) {
	if got, want := PDFName("", KindResume, "Analyst", "Acme"), "resume-analyst-acme.pdf"; got != want {
		t.Errorf("empty name: got %q, want %q", got, want)
	}
	if got, want := PDFName("Mira", KindCoverLetter, "", "Acme"), "Mira-cover-letter-acme.pdf"; got != want {
		t.Errorf("empty position: got %q, want %q", got, want)
	}
	long := strings.Repeat("Abcdefghij ", 40)
	if got := PDFName(long, KindCoverLetter, long, long); len(got) >= 255 {
		t.Errorf("worst-case name is %d bytes, want < 255", len(got))
	}
}

func TestParseKind(t *testing.T) {
	for _, s := range []string{"resume", "cover-letter"} {
		if k, err := ParseKind(s); err != nil || string(k) != s {
			t.Errorf("ParseKind(%q) = %q, %v", s, k, err)
		}
	}
	for _, s := range []string{"", "cv", "Resume", "cover_letter"} {
		_, err := ParseKind(s)
		if !errors.Is(err, ErrKind) || !strings.Contains(err.Error(), "resume|cover-letter") {
			t.Errorf("ParseKind(%q) err = %v, want ErrKind naming resume|cover-letter", s, err)
		}
	}
}
