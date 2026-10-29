package document

import (
	"reflect"
	"testing"
)

type matchCase struct {
	name    string
	doc     string
	answers []string
	targets []string
	want    []string // "kind:text" of each untraceable claim; nil = passes
}

func runMatchCases(t *testing.T, cases []matchCase) {
	t.Helper()
	p := testProfile(t)
	for _, c := range cases {
		r := Verify([]byte(c.doc), Sources{Profile: p, Answers: c.answers, Targets: c.targets})
		var got []string
		for _, cl := range r.Untraceable {
			got = append(got, cl.Kind+":"+cl.Text)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
		if r.Claims == 0 && c.want == nil && c.name != "non-factual prose" {
			t.Errorf("%s: no claims extracted", c.name)
		}
	}
}

// Exact and fuzzy matches: spelling of months, separators, thousands,
// number words, case and diacritics.
func TestVerifyMatches(t *testing.T) {
	runMatchCases(t, []matchCase{
		{name: "exact month-year", doc: "- Joined in Mar 2023."},
		{name: "Jun vs June", doc: "- Visiting from June 2022."},
		{name: "Sep vs September", doc: "- Analyst from September 2019."},
		{name: "Sept vs Sep and September", doc: "- Studied from Sep 2014 and September 2014."},
		{name: "bare year", doc: "- Graduated in 2019."},
		{name: "range in prose", doc: "- Researcher, Jan 2022 – Jun 2022."},
		{name: "40000 vs 40,000", doc: "- Linked data for 40000 patients."},
		{name: "digits vs number words", doc: "- Mentor 2 junior analysts in 3 trials."},
		{name: "case", doc: "- Analyst at NORTHWIND CLINICAL DATA."},
		{name: "diacritics", doc: "- Based in Utrécht and Bérgen."},
		{name: "diacritics in answers", doc: "- Advised Fundacao Exemplo.", answers: []string{"Advised Fundação Exemplo"}},
		{name: "possessive", doc: "- Built Fjordline's survival pipeline."},
		{name: "hyphenated name", doc: "- Worked with Halvorsen and Utrecht-based teams."},
		{name: "terms", doc: "- Ran R labs; Bayesian models; ORCID linked."},
		{name: "link", doc: "See [GitHub](https://github.com/example/) for code."},
		{name: "work fields", doc: "### Nowcasting outbreaks with delayed reporting\n" +
			"Proceedings of the Imaginary Epidemiology Conference\nExample-Halvorsen M and Smith A\n2022"},
		{name: "entry with hyphen range", doc: "### Data Analyst\nNorthwind Clinical Data\nSeptember 2019 - February 2023"},
		{name: "entry fields joined", doc: "### Visiting Researcher\n**Institute for Imaginary Epidemiology** | Leiden, Netherlands\nJan 2022 – June 2022"},
		{name: "letter conventions", doc: "Dear Hiring Manager,\n\nKind regards,\n\nMira Example-Halvorsen"},
		{name: "targets named", doc: "I want to join the Example Health Alliance.", targets: []string{"Example Health Alliance"}},
		{name: "answer number", doc: "My notice period is 1 month.", answers: []string{"1 month"}},
		{name: "answer ISO date", doc: "I can start in January 2027.", answers: []string{"2027-01-01"}},
		{name: "non-factual prose", doc: "I am passionate about open science and careful work."},
	})
}

// False positives: inventions are caught even when each word exists.
func TestVerifyCatchesInventions(t *testing.T) {
	runMatchCases(t, []matchCase{
		{name: "invented employer", doc: "- Consulted for Acme Pharma on trials.", want: []string{"term:Acme Pharma"}},
		{name: "recombined name", doc: "- Analyst at Fjordline Clinical Institute.", want: []string{"term:Fjordline Clinical Institute"}},
		{name: "invented skill", doc: "- Deployed models with PyTorch.", want: []string{"term:PyTorch"}},
		{name: "skill at sentence start", doc: "- PyTorch pipelines for trials.", want: []string{"term:PyTorch"}},
		{name: "acronym", doc: "- Used SAS daily.", want: []string{"term:SAS"}},
		{name: "symbol term", doc: "- wrote C++ tools", want: []string{"term:C++"}},
		{name: "prose skill", doc: "I am passionate about Kubernetes.", want: []string{"term:Kubernetes"}},
		{name: "after an abbreviation", doc: "Dear Dr. Jansen,", want: []string{"term:Jansen"}},
		{name: "wrapped line", doc: "Pipelines in R and\nKubernetes.", want: []string{"term:Kubernetes"}},
		{name: "number", doc: "- Ran labs for 80 students.", want: []string{"number:80"}},
		{name: "number word", doc: "- Mentor five junior analysts.", want: []string{"number:five"}},
		{name: "percentage", doc: "- Cut reporting time by 30%.", want: []string{"number:30%"}},
		{name: "month not in history", doc: "- Started my MSc in Sep 2017.", want: []string{"date:Sep 2017"}},
		{name: "year not in history", doc: "- Analyst since 2016.", want: []string{"date:2016"}},
		{name: "day date", doc: "I can start on 1 January 2027.", want: []string{"date:1 January 2027"}},
		{name: "link", doc: "Code at https://github.com/mira.", want: []string{"term:https://github.com/mira"}},
		{name: "email", doc: "Write to mira@example.org.", want: []string{"term:mira@example.org"}},
		{name: "target without targets", doc: "I want to join the Example Health Alliance.",
			want: []string{"term:Example Health Alliance"}},
		{name: "targets do not vouch for skills", doc: "## Skills\n- Biostatistics, Kubernetes",
			targets: []string{"Kubernetes Engineer"}, want: []string{"skill:Kubernetes"}},
		{name: "answer needed", doc: "My notice period is 1 month.", want: []string{"number:1"}},
		{name: "swapped dates", doc: "### Visiting Researcher\nInstitute for Imaginary Epidemiology\nMar 2023 – Present",
			want: []string{"entry:Mar 2023 – Present"}},
		{name: "unknown entry", doc: "### Head of Data\nNorthwind Clinical Data\nSep 2019 – Feb 2023",
			want: []string{"entry:Head of Data"}},
		{name: "frontmatter value", doc: "---\ntitle: Kubernetes Engineer\nauthor: Mira Example-Halvorsen\n---\nBody.",
			want: []string{"term:Kubernetes Engineer"}},
	})
}
