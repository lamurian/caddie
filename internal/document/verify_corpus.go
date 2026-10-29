package document

import (
	"regexp"
	"strings"

	"github.com/lamurian/caddie/internal/profile"
)

// verifyCorpus is everything a draft may claim: the profile and the
// answers in scope, plus the posting's company and title (targets), which
// vouch for named terms only.
type verifyCorpus struct {
	texts   []string // " word word " phrases from the profile and answers
	targets []string // the same, from the targets
	urls    map[string]bool
	dates   map[string]bool // "YYYY", "YYYY-MM", "YYYY-MM-DD"
	numbers map[string]bool
	records []verifyRecord
}

// verifyRecord is one history entry or work, folded for comparison.
type verifyRecord struct {
	work       bool
	title      string
	fields     map[string]bool // organisation, location or venue
	authors    string
	start, end string
}

// verifyURLRe finds URLs, e-mail addresses and bare "host.tld/path" links.
var verifyURLRe = regexp.MustCompile(`(?i)\b(?:https?://|www\.)[^\s<>()\[\]]+` +
	`|[\w.+-]+@[\w-]+(?:\.[\w-]+)+` +
	`|\b[a-z0-9-]+(?:\.[a-z0-9-]+)*\.(?:com|org|net|io|dev|edu|gov|int|eu|ai|app|me)/[^\s<>()\[\]]*`)

func newVerifyCorpus(src Sources) *verifyCorpus {
	c := &verifyCorpus{urls: map[string]bool{}, dates: map[string]bool{}, numbers: map[string]bool{}}
	p := src.Profile
	id := p.Identity
	c.add(id.Name, id.Nationality, id.Location, p.About)
	c.add(id.WorkAuthorization...)
	for k, u := range id.Links {
		c.add(k, u)
	}
	for _, l := range id.Languages {
		c.add(l.Name + " " + l.Level)
	}
	for _, s := range p.Sections {
		c.add(s.Heading)
		for _, e := range s.Entries {
			c.addEntry(e)
		}
		for _, w := range s.Works {
			c.addWork(w)
		}
	}
	c.add(src.Answers...)
	for _, t := range src.Targets {
		c.targets = append(c.targets, verifyPhrases(stripURLs(t, nil))...)
	}
	return c
}

func (c *verifyCorpus) addEntry(e profile.Entry) {
	c.add(e.Title, e.Organisation, e.Location)
	c.add(e.Bullets...)
	for _, d := range []string{e.Start, e.End} {
		if d != "present" {
			addVerifyDate(c.dates, d)
		}
	}
	c.records = append(c.records, verifyRecord{title: foldField(e.Title), start: e.Start, end: e.End,
		fields: map[string]bool{foldField(e.Organisation): true, foldField(e.Location): true}})
}

func (c *verifyCorpus) addWork(w profile.Work) {
	c.add(w.Title, w.Venue, strings.Join(w.Authors, ", "))
	c.add(w.Bullets...)
	addVerifyDate(c.dates, w.Date)
	c.records = append(c.records, verifyRecord{work: true, title: foldField(w.Title),
		fields: map[string]bool{foldField(w.Venue): true}, authors: authorsKey(w.Authors), start: w.Date})
}

// add indexes texts for every claim kind: URLs, dates, numbers, phrases.
func (c *verifyCorpus) add(texts ...string) {
	for _, t := range texts {
		rest := stripURLs(t, func(u string) { c.urls[normURL(u)] = true })
		c.texts = append(c.texts, verifyPhrases(rest)...)
		dates, plain := scanVerifyDates(rest)
		for _, d := range dates {
			addVerifyDate(c.dates, d.norm)
		}
		for _, w := range splitVerifyWords(plain) {
			if n, ok := verifyNumber(w); ok {
				c.numbers[n] = true
			}
		}
	}
}

// verifyPhrases returns t as " key key " and, when it has hyphens, also
// with hyphens as spaces, so "Halvorsen" matches "Example-Halvorsen".
func verifyPhrases(t string) []string {
	var keys []string
	for _, w := range splitVerifyWords(t) {
		if w.key != "" {
			keys = append(keys, w.key)
		}
	}
	p := " " + strings.Join(keys, " ") + " "
	if h := strings.ReplaceAll(p, "-", " "); h != p {
		return []string{p, h}
	}
	return []string{p}
}

// hasPhrase reports whether the folded words of phrase appear, in order
// and as whole words, in one text. Targets count only when asked.
func (c *verifyCorpus) hasPhrase(phrase string, targets bool) bool {
	p := " " + phrase + " "
	if p == "  " {
		return true
	}
	for _, t := range c.texts {
		if strings.Contains(t, p) {
			return true
		}
	}
	for _, t := range c.targets {
		if targets && strings.Contains(t, p) {
			return true
		}
	}
	return false
}

// stripURLs replaces URLs and e-mail addresses with a comma, passing each
// one (edge punctuation trimmed) to found when it is not nil.
func stripURLs(s string, found func(string)) string {
	return verifyURLRe.ReplaceAllStringFunc(s, func(u string) string {
		if found != nil {
			found(strings.TrimRight(u, ".,;:!?'\")"))
		}
		return " , "
	})
}

// normURL compares links without scheme, "www.", case or trailing slash.
func normURL(u string) string {
	u = strings.TrimRight(foldClaim(u), ".,;:!?'\")")
	for _, p := range []string{"mailto:", "https://", "http://", "www."} {
		u = strings.TrimPrefix(u, p)
	}
	return strings.TrimRight(u, "/")
}

// foldField folds a heading or field line: emphasis and edge punctuation
// go, spaces collapse.
func foldField(s string) string {
	s = strings.Trim(strings.TrimSpace(s), "*_`\\ ")
	return strings.Join(strings.Fields(foldClaim(s)), " ")
}

func authorsKey(authors []string) string {
	keys := make([]string, len(authors))
	for i, a := range authors {
		keys[i] = foldField(a)
	}
	return strings.Join(keys, "|")
}
