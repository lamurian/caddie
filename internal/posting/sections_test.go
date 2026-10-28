package posting

import (
	"reflect"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := map[string]int{
		"What we offer":           Perks,
		"Benefits & perks":        Perks,
		"What you'll bring":       Requirements,
		"Qualifications":          Requirements,
		"Nice to have":            Requirements,
		"Key responsibilities":    Responsibilities,
		"What you will do":        Responsibilities,
		"What You’ll Do":          Responsibilities,
		"Why Acme":                Perks,
		"About Acme":              Description,
		"Company profile? Maybe.": Description,
	}
	for heading, want := range cases {
		if got, _ := classify(heading); got != want {
			t.Errorf("classify(%q) = %s, want %s", heading, SectionNames[got], SectionNames[want])
		}
	}
}

func TestExtractMarkdown(t *testing.T) {
	s := Extract("<h1>Analyst</h1><p>Intro.</p><h2>Responsibilities</h2><ul><li>One</li><li># Two</li></ul>"+
		"<h2>About us</h2><p>We are fictional.</p>", "analyst")
	want := "## Description\n\nIntro.\n\n**About us**\n\nWe are fictional.\n\n" +
		"## Responsibilities\n\n- One\n- \\# Two\n\n" +
		"## Requirements\n\n" + notStated + "\n\n## Perks and benefits\n\n" + notStated + "\n\n"
	if got := s.Markdown(); got != want {
		t.Errorf("Markdown:\n%s\nwant:\n%s", got, want)
	}
	if got := s.Missing(); !reflect.DeepEqual(got, []string{"Requirements", "Perks and benefits"}) {
		t.Errorf("Missing = %v", got)
	}
}
