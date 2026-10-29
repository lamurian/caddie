package document

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/profile"
)

// requirements wraps lines in a minimal posting with frontmatter and
// neighbouring sections whose words must not leak in.
func requirements(lines ...string) []byte {
	return []byte("---\nslug: x\n---\n\n## Description\n\n- Python everywhere\n\n## Requirements\n\n" +
		strings.Join(lines, "\n") + "\n\n## Perks and benefits\n\n- Kubernetes budget\n")
}

func phraseTexts(posting []byte) []string {
	ps, _ := requirementPhrases(posting)
	var out []string
	for _, p := range ps {
		out = append(out, p.text)
	}
	return out
}

func TestRequirementPhrases(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []string
	}{
		{"filler prefix", "- Experience with Bayesian methods", []string{"bayesian methods"}},
		{"years and lead-in", "- 5+ years of experience in survival analysis.", []string{"survival analysis"}},
		{"year range", "- 3-5 years clinical trials", []string{"clinical trials"}},
		{"stacked filler", "- Strong programming skills in R and Python", []string{"r", "python"}},
		{"separators", "- SQL, Spark; Stan / JAGS (PyMC) or Julia & C++",
			[]string{"sql", "spark", "stan", "jags", "pymc", "julia", "c++"}},
		{"label before colon", "* Languages: Go and Rust", []string{"go", "rust"}},
		{"filler suffix", "1. Communication skills; SQL is a plus", []string{"communication", "sql"}},
		{"abbreviation", "- Bayesian tools (e.g. Stan)", []string{"bayesian tools", "stan"}},
		{"noise only", "- PhD in statistics or a related field", []string{"phd in statistics"}},
		{"dedupe by tokens", "- R\n- r, Python\n- python", []string{"r", "python"}},
		{"dedupe plural", "- Mixed models\n- mixed model", []string{"mixed models"}},
		{"bold markup", "- **Kotlin** and `Go`", []string{"kotlin", "go"}},
		{"labels skipped", "**What you bring**\nNice to have:\n### Extra\n- Rust", []string{"rust"}},
		{"too long", "- Explain complex statistical results to non-technical audiences", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := phraseTexts(requirements(tt.line)); !slices.Equal(got, tt.want) {
				t.Errorf("phrases = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRequirementNotices(t *testing.T) {
	tests := []struct {
		name    string
		posting string
		want    string
	}{
		{"no section", "---\nslug: x\n---\n\n## Description\n\n- R\n", "posting has no ## Requirements section"},
		{"empty section", "## Requirements\n\n**What you bring**\n\n## Perks and benefits\n",
			"the ## Requirements section is empty"},
		{"long phrase", "## Requirements\n- Ability to explain one two three four five six\n",
			`skipped (over 5 words): "explain one two three four five six"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps, notices := requirementPhrases([]byte(tt.posting))
			if len(ps) != 0 || !slices.Equal(notices, []string{tt.want}) {
				t.Errorf("phrases %v, notices %q, want %q", ps, notices, tt.want)
			}
		})
	}
}

func TestKeywordMatching(t *testing.T) {
	tests := []struct {
		keyword, text string
		want          bool
	}{
		{"r", "Pipelines in R.", true},
		{"r", "Ran weekly labs", false},
		{"r", "R&D budget", false},
		{"time-to-event data", "time to event data", true},
		{"mixed models", "a mixed-model analysis", true},
		{"msc in biostatistics", "### MSc Biostatistics", true},
		{"cafe analytics", "Café Analytics", true},
		{"node.js", "Node.js services", true},
		{"survival analysis", "survival. Analysis", false},
		{"survival analysis", "- survival\n- analysis", false},
		{"survival analysis", "built a survival\n  analysis pipeline", true},
		{"c++", "C++ and C", true},
		{"c", "C++ only", false},
	}
	for _, tt := range tests {
		got := containsRun(markdownTokens(tt.text), contentTokens(tt.keyword))
		if got != tt.want {
			t.Errorf("%q in %q = %v, want %v", tt.keyword, tt.text, got, tt.want)
		}
	}
}

func TestKeywordStatuses(t *testing.T) {
	prof := profile.Profile{About: "I use R and a survival model.",
		Identity: profile.Identity{Languages: []profile.Language{{Name: "Dutch"}}}}
	posting := requirements("- R, SAS, Dutch, Kubernetes, survival models")
	r := Keywords(posting, []byte("---\ntitle: x\n---\nR and SAS; survival models\n"), prof)
	want := map[string]string{"r": KeywordCovered, "sas": KeywordUnbacked, "dutch": KeywordMissing,
		"kubernetes": KeywordUnsupported, "survival models": KeywordCovered}
	if r.Covered != 2 || r.Total != 5 || len(r.Notices) != 1 {
		t.Errorf("covered %d/%d, notices %q", r.Covered, r.Total, r.Notices)
	}
	for _, k := range r.Keywords {
		if want[k.Text] != k.Status {
			t.Errorf("%q: status %q, want %q", k.Text, k.Status, want[k.Text])
		}
	}
	data, err := json.Marshal(r.Keywords[0])
	if err != nil || string(data) != `{"keyword":"r","in_draft":true,"in_history":true,"status":"covered"}` {
		t.Errorf("json = %s, %v", data, err)
	}
}

func TestKeywordsNoRequirements(t *testing.T) {
	r := Keywords([]byte("## Description\n\nR\n"), []byte("R"), profile.Profile{})
	want := "keywords: 0/0 covered, 0 missing, 0 unsupported, 0 unbacked\n" +
		"note: posting has no ## Requirements section\n"
	if r.Text() != want || r.Keywords == nil {
		t.Errorf("text = %q, keywords %v", r.Text(), r.Keywords)
	}
}
