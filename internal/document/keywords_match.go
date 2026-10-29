package document

import (
	"regexp"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/lamurian/caddie/internal/profile"
)

// matchStopwords are ignored on both sides when matching, so
// "MSc in Biostatistics" matches "MSc Biostatistics".
var matchStopwords = map[string]bool{"a": true, "an": true, "the": true, "of": true, "in": true,
	"on": true, "for": true, "to": true, "with": true, "at": true, "by": true}

// phraseBreaks end a phrase: a keyword never matches across them.
const phraseBreaks = ".,;:!?()[]{}\"|/"

// blockStart marks a Markdown line that starts a new block.
var blockStart = regexp.MustCompile(`^(?:[-*+#>|]|\d+[.)]\s)`)

// historyText flattens the parts of a profile that can support a keyword:
// about, language names, section headings, entry titles, organisations and
// bullets, and work titles, venues and bullets. Each part is its own block.
func historyText(p profile.Profile) string {
	parts := []string{p.About}
	for _, l := range p.Identity.Languages {
		parts = append(parts, l.Name)
	}
	for _, s := range p.Sections {
		parts = append(parts, s.Heading)
		for _, e := range s.Entries {
			parts = append(append(parts, e.Title, e.Organisation), e.Bullets...)
		}
		for _, w := range s.Works {
			parts = append(append(parts, w.Title, w.Venue), w.Bullets...)
		}
	}
	return strings.Join(parts, "\n\n")
}

// markdownTokens tokenises Markdown text. "" marks a phrase boundary:
// break punctuation, a blank line, or a line that starts a new block
// (bullet, heading, quote, table row, numbered item). A wrapped line
// continues its paragraph or bullet.
func markdownTokens(s string) []string {
	var out []string
	prevBlank := true
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if prevBlank || blockStart.MatchString(t) {
			out = append(out, "")
		}
		out = append(out, contentTokens(t)...)
		prevBlank = t == ""
	}
	return out
}

// contentTokens folds s (lowercase, diacritics removed), splits it into
// words of letters, digits, "+", "#" and "&", drops matchStopwords and
// stems simple plurals. A hyphen, space or inner dot ("node.js") separates
// words; phraseBreaks add a "" boundary; apostrophes are dropped.
func contentTokens(s string) []string {
	folded, _, err := transform.String(transform.Chain(norm.NFD,
		runes.Remove(runes.In(unicode.Mn)), norm.NFC), strings.ToLower(s))
	if err != nil {
		folded = strings.ToLower(s)
	}
	rs := []rune(folded)
	var out []string
	var word strings.Builder
	flush := func() {
		if w := word.String(); w != "" && !matchStopwords[w] {
			out = append(out, stemPlural(w))
		}
		word.Reset()
	}
	for i, r := range rs {
		switch {
		case isKeywordRune(r):
			word.WriteRune(r)
		case r == '\'' || r == '’':
		case strings.ContainsRune(phraseBreaks, r) && !innerDot(rs, i):
			flush()
			out = append(out, "")
		default:
			flush()
		}
	}
	flush()
	return out
}

func isKeywordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '+' || r == '#' || r == '&'
}

// innerDot reports a dot between two word runes, as in "node.js".
func innerDot(rs []rune, i int) bool {
	return rs[i] == '.' && i > 0 && i+1 < len(rs) && isKeywordRune(rs[i-1]) && isKeywordRune(rs[i+1])
}

// stemPlural drops a final "s" from words over three letters, except
// "-ss", "-us" and "-is" ("class", "status", "analysis").
func stemPlural(w string) string {
	if len(w) <= 3 || !strings.HasSuffix(w, "s") {
		return w
	}
	for _, keep := range []string{"ss", "us", "is"} {
		if strings.HasSuffix(w, keep) {
			return w
		}
	}
	return w[:len(w)-1]
}

// containsRun reports whether needle occurs as a contiguous run of tokens
// in hay, so matches are whole words and never cross a boundary.
func containsRun(hay, needle []string) bool {
	if len(needle) == 0 {
		return false
	}
	for i := 0; i+len(needle) <= len(hay); i++ {
		if slices.Equal(hay[i:i+len(needle)], needle) {
			return true
		}
	}
	return false
}
