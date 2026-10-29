package document

import (
	"fmt"
	"regexp"
	"strings"
)

// maxKeywordWords: a longer phrase is prose, not a keyword; it is skipped
// with a notice so the agent can still read it in the posting.
const maxKeywordWords = 5

// requirementPhrase is one extracted keyword: display text and the
// content tokens it is matched by.
type requirementPhrase struct {
	text   string
	tokens []string
}

var (
	// requirementSplitter turns every phrase separator into "\x00".
	requirementSplitter = strings.NewReplacer(",", "\x00", ";", "\x00", "/", "\x00", "(", "\x00",
		")", "\x00", "|", "\x00", ". ", "\x00", " and ", "\x00", " or ", "\x00", " & ", "\x00")
	// abbreviations would otherwise split a phrase at their dots.
	abbreviations = strings.NewReplacer("e.g.", " ", "i.e.", " ", "etc.", " ")
	// yearsPrefix: "5+ years of", "3-5 years", "2 years".
	yearsPrefix = regexp.MustCompile(`^\d+(?:\s*[-–]\s*\d+)?\s*\+?\s*years?\b(?:\s+of\b)?\s*`)
	// leadIn: everything up to "skills in", "experience with", "knowledge of", ….
	leadIn = regexp.MustCompile(`^.*?\b(?:skills?|experience|expertise|proficiency|proficient|` +
		`knowledge|familiarity|familiar|background|understanding)\s+(?:in|with|of|using)\s+`)
	// fillerPrefix: words that qualify a requirement without naming it.
	fillerPrefix = regexp.MustCompile(`^(?:you have|you are|you bring|must have|nice to have|an?|the|` +
		`at least|proven|demonstrated|strong|excellent|solid|very good|good|deep|extensive|hands-on|` +
		`advanced|fluent|fluency in|ability to|able to|experience|degree in|such as|including)\s+`)
	// fillerSuffix: trailing words that say how much a requirement matters.
	fillerSuffix = regexp.MustCompile(`\s+(?:skills?|experience|knowledge|(?:is )?an? (?:plus|advantage)|` +
		`(?:is )?(?:preferred|required|desirable))$`)
)

// noiseWords never make a keyword on their own (compared after folding).
var noiseWords = map[string]bool{"related": true, "relevant": true, "similar": true, "equivalent": true,
	"field": true, "other": true, "preferred": true, "required": true, "desirable": true, "plus": true,
	"advantage": true, "etc": true, "is": true, "are": true, "be": true, "and": true, "or": true}

// requirementPhrases extracts the de-duplicated keywords of the posting's
// "## Requirements" section, with notices for skipped phrases.
func requirementPhrases(posting []byte) ([]requirementPhrase, []string) {
	lines, found := requirementsLines(posting)
	if !found {
		return nil, []string{"posting has no ## Requirements section"}
	}
	if len(lines) == 0 {
		return nil, []string{"the ## Requirements section is empty"}
	}
	var out []requirementPhrase
	var notices []string
	seen := map[string]bool{}
	for _, l := range lines {
		for _, raw := range splitRequirement(l) {
			p, keep := toRequirementPhrase(raw)
			key := strings.Join(p.tokens, " ")
			switch {
			case !keep || seen[key]:
			case len(strings.Fields(p.text)) > maxKeywordWords:
				notices = append(notices, fmt.Sprintf("skipped (over %d words): %q", maxKeywordWords, p.text))
			default:
				out = append(out, p)
			}
			seen[key] = true
		}
	}
	return out, notices
}

// requirementsLines returns the content lines between "## Requirements"
// and the next "#"/"##" heading, without bullet markers or label lines.
func requirementsLines(posting []byte) (lines []string, found bool) {
	in := false
	for _, raw := range strings.Split(string(keywordsBody(posting)), "\n") {
		t := strings.TrimSpace(raw)
		if strings.HasPrefix(t, "# ") || strings.HasPrefix(t, "## ") {
			if in {
				break
			}
			in = strings.EqualFold(strings.TrimSpace(strings.TrimLeft(t, "#")), "requirements")
			found = found || in
			continue
		}
		if in && isRequirementLine(t) {
			lines = append(lines, stripRequirementMarkup(t))
		}
	}
	return lines, found
}

// isRequirementLine rejects blanks, sub-headings and labels such as
// "**What you bring**" or "Nice to have:".
func isRequirementLine(t string) bool {
	bold := strings.HasPrefix(t, "**") && strings.HasSuffix(t, "**")
	return t != "" && !strings.HasPrefix(t, "#") && !bold && !strings.HasSuffix(t, ":")
}

var listMarker = regexp.MustCompile(`^(?:[-*+]|\d+[.)])\s+`)

func stripRequirementMarkup(t string) string {
	t = listMarker.ReplaceAllString(t, "")
	return strings.NewReplacer("**", "", "__", "", "`", "").Replace(t)
}

// splitRequirement lowercases a line, keeps only the text after a
// "Label:" prefix and splits it into raw phrases.
func splitRequirement(line string) []string {
	s := abbreviations.Replace(strings.ToLower(line))
	if i := strings.Index(s, ":"); i >= 0 && strings.TrimSpace(s[i+1:]) != "" {
		s = s[i+1:]
	}
	return strings.Split(requirementSplitter.Replace(s), "\x00")
}

// toRequirementPhrase strips filler until nothing changes; keep is false
// when no content word is left.
func toRequirementPhrase(raw string) (requirementPhrase, bool) {
	s := trimPhrase(raw)
	for prev := ""; s != prev; {
		prev = s
		s = yearsPrefix.ReplaceAllString(s, "")
		s = leadIn.ReplaceAllString(s, "")
		s = fillerPrefix.ReplaceAllString(s, "")
		s = trimPhrase(fillerSuffix.ReplaceAllString(s, ""))
	}
	p := requirementPhrase{text: s}
	for _, t := range contentTokens(s) {
		if t != "" {
			p.tokens = append(p.tokens, t)
		}
	}
	for _, t := range p.tokens {
		if !noiseWords[t] {
			return p, true
		}
	}
	return p, false
}

func trimPhrase(s string) string {
	return strings.Trim(strings.Join(strings.Fields(s), " "), " .!?:'\"*-–")
}
