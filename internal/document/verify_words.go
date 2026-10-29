package document

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// verifyWord is one word of a line, as Verify reads it.
type verifyWord struct {
	raw  string // as written, edge punctuation trimmed
	key  string // folded and without 's, 'm…: what is matched
	open bool   // punctuation before the word: a phrase starts here
	stop bool   // punctuation after the word: a phrase ends here
	end  bool   // the word ends a sentence (. ! ?)
}

const (
	verifyEdgeLeft  = "()[]{}\"'*_`<“‘«~≈$€£-"
	verifyEdgeRight = "()[]{}\"'*_`>.,;:!?”’»…-\\"
)

var (
	// verifyFoldExtra covers letters Unicode does not decompose.
	verifyFoldExtra = strings.NewReplacer("ø", "o", "æ", "ae", "ß", "ss", "ł", "l",
		"đ", "d", "ı", "i", "œ", "oe", "’", "'", "‘", "'")
	verifyContraction = regexp.MustCompile(`'(s|m|re|ve|d|ll)$`)
	verifyNumberRe    = regexp.MustCompile(`^\d+(?:[.,]\d+)*[%+]?$`)
	verifyThousandsRe = regexp.MustCompile(`^\d{1,3}(?:,\d{3})+(?:\.\d+)?$`)
	// verifyNumberWords are matched as digits on both sides. "one" is left
	// out: it is far more often a pronoun than a quantity.
	verifyNumberWords = map[string]string{"two": "2", "three": "3", "four": "4",
		"five": "5", "six": "6", "seven": "7", "eight": "8", "nine": "9",
		"ten": "10", "eleven": "11", "twelve": "12"}
)

// foldClaim lowercases s and strips diacritics: "Utrécht" and "UTRECHT"
// fold alike.
func foldClaim(s string) string {
	s = strings.ToLower(s)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	if out, _, err := transform.String(t, s); err == nil {
		s = out
	}
	return verifyFoldExtra.Replace(s)
}

func isVerifySep(r rune) bool {
	return unicode.IsSpace(r) || strings.ContainsRune("/–—|·•", r)
}

// splitVerifyWords splits a line on spaces, slashes, dashes and bars.
// Hyphens and dots inside a word are kept: "Example-Halvorsen", "Node.js".
func splitVerifyWords(s string) []verifyWord {
	var out []verifyWord
	for _, f := range strings.FieldsFunc(s, isVerifySep) {
		core := strings.TrimRight(f, verifyEdgeRight)
		trail := f[len(core):]
		raw := strings.TrimLeft(core, verifyEdgeLeft)
		w := verifyWord{raw: raw, open: len(raw) < len(core), stop: trail != "",
			end: strings.ContainsAny(trail, ".!?…")}
		w.key = verifyContraction.ReplaceAllString(foldClaim(raw), "")
		out = append(out, w)
	}
	return out
}

// verifyNumber returns the normalised quantity a word states: "40,000" →
// "40000", "30%" → "30", "07" → "7", "three" → "3".
func verifyNumber(w verifyWord) (string, bool) {
	if d, ok := verifyNumberWords[w.key]; ok {
		return d, true
	}
	if !verifyNumberRe.MatchString(w.raw) {
		return "", false
	}
	n := strings.TrimRight(w.raw, "%+")
	if verifyThousandsRe.MatchString(n) {
		n = strings.ReplaceAll(n, ",", "")
	}
	if t := strings.TrimLeft(n, "0"); t != "" && t[0] != '.' && t[0] != ',' {
		n = t
	}
	return n, true
}

// verifyAllow are words that may be capitalised without being claims:
// letter conventions, calendar words and abbreviations.
var verifyAllow = wordSet(`i dear sincerely regards kind best warm warmest
	yours faithfully truly respectfully hiring manager team committee sir madam
	mr mrs ms dr re to whom it concern present resume cv cover letter
	curriculum vitae e.g i.e etc vs january february march april may june july
	august september october november december jan feb mar apr jun jul aug sep
	sept oct nov dec monday tuesday wednesday thursday friday saturday sunday`)

// verifyAbbrev end with a dot without ending the sentence: the capital
// after "Dr." is still a claim.
var verifyAbbrev = wordSet("dr mr mrs ms prof st e.g i.e cf vs approx")

func wordSet(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}

// isTermShaped reports whether a word names something: ALL-CAPS, an
// inner capital, a digit or a symbol anywhere, or a capital that does not
// start a sentence. Allowed words and plain lowercase words are not terms.
func isTermShaped(w verifyWord, initial bool) bool {
	if w.key == "" || verifyAllow[w.key] {
		return false
	}
	var letters, upper, lower int
	inner, symbol, digit := false, false, false
	var prev rune
	for _, r := range w.raw {
		switch {
		case unicode.IsUpper(r):
			upper++
			letters++
			inner = inner || unicode.IsLower(prev)
		case unicode.IsLetter(r):
			lower++
			letters++
		case unicode.IsDigit(r):
			digit = true
		case r != '-' && r != '\'' && r != '’':
			symbol = true
		}
		prev = r
	}
	switch {
	case letters == 0:
		return digit // "12(3)": numbers are handled before this
	case upper >= 2 && lower == 0, inner, digit, symbol:
		return true
	}
	return !initial && unicode.IsUpper([]rune(w.raw)[0])
}
