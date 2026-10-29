package profile

import (
	"reflect"
	"testing"
)

const fixture = "../../testdata/history"

func section(t *testing.T, p Profile, id string) Section {
	t.Helper()
	for _, s := range p.Sections {
		if s.ID == id {
			return s
		}
	}
	t.Fatalf("section %q not found", id)
	return Section{}
}

func TestLoadFixtureIdentity(t *testing.T) {
	p, err := Load(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if p.Identity.Name != "Mira Example-Halvorsen" || len(p.Identity.Languages) != 3 {
		t.Errorf("identity: %+v", p.Identity)
	}
	if p.Identity.Links["orcid"] == "" || p.About == "" {
		t.Errorf("links or about missing: %+v", p)
	}
	ids := []string{}
	for _, s := range p.Sections {
		ids = append(ids, s.ID)
	}
	want := []string{"consultancies", "education", "employment", "presentation", "publication", "research", "teaching"}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("sections = %v, want %v (README.md and about.md skipped)", ids, want)
	}
}

func TestLoadFixtureEntries(t *testing.T) {
	p, err := Load(fixture)
	if err != nil {
		t.Fatal(err)
	}
	emp := section(t, p, "employment")
	want := Entry{
		Title: "Senior Biostatistician", Organisation: "Fjordline Health Research",
		Location: "Utrecht, Netherlands", Start: "2023-03", End: "present",
		Bullets: []string{
			"Lead the statistical analysis plan for three multi-centre trials",
			"Built a survival analysis pipeline in R that cut reporting time from two weeks to two days",
			"Mentor two junior analysts",
		},
	}
	if !reflect.DeepEqual(emp.Entries[0], want) {
		t.Errorf("present entry:\n got %+v\nwant %+v", emp.Entries[0], want)
	}
	noLoc := emp.Entries[1]
	if noLoc.Location != "" || noLoc.Organisation != "Northwind Clinical Data" || noLoc.End != "2023-02" {
		t.Errorf("entry without location: %+v", noLoc)
	}
	edu := section(t, p, "education")
	if edu.Entries[0].Start != "2017" || edu.Entries[1].Start != "2014-09" || edu.Entries[1].Bullets != nil {
		t.Errorf("education dates: %+v", edu.Entries)
	}
	if c := section(t, p, "consultancies").Entries[0]; c.Start != "2021-08" || c.End != "2021-12" {
		t.Errorf("full month name / hyphen: %+v", c)
	}
}

func TestLoadFixtureWorks(t *testing.T) {
	p, err := Load(fixture)
	if err != nil {
		t.Fatal(err)
	}
	pubs := section(t, p, "publication").Works
	if len(pubs) != 2 {
		t.Fatalf("publications: %+v", pubs)
	}
	wantAuthors := []string{"Example-Halvorsen M", "Doe J", "Roe R", "Poe P"}
	if !reflect.DeepEqual(pubs[0].Authors, wantAuthors) || pubs[0].Date != "2024-03" {
		t.Errorf("publication 1: %+v", pubs[0])
	}
	if pubs[0].Venue != "Journal of Fictional Statistics, 12(3), 45–67" || len(pubs[0].Bullets) != 1 {
		t.Errorf("publication venue/bullets: %+v", pubs[0])
	}
	if len(pubs[1].Authors) != 2 || pubs[1].Date != "2022" {
		t.Errorf("publication 2: %+v", pubs[1])
	}
	pres := section(t, p, "presentation").Works
	if len(pres) != 1 || pres[0].Date != "2023-11" || pres[0].Authors[0] != "Example-Halvorsen M" {
		t.Errorf("presentation: %+v", pres)
	}
}

func TestParseRange(t *testing.T) {
	cases := []struct{ in, start, end string }{
		{"Aug 2022 – Jun 2023", "2022-08", "2023-06"},
		{"Aug 2022 – Present", "2022-08", "present"},
		{"aug. 2022-present", "2022-08", "present"},
		{"2019 — 2021", "2019", "2021"},
		{"May 2020 – 2020", "2020-05", "2020"},
	}
	for _, c := range cases {
		s, e, err := ParseRange(c.in)
		if err != nil || s != c.start || e != c.end {
			t.Errorf("%q: got %q %q %v", c.in, s, e, err)
		}
	}
}
