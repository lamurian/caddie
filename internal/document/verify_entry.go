package document

import (
	"strings"

	"github.com/lamurian/caddie/internal/profile"
)

// verifyLine is one numbered line of a draft.
type verifyLine struct {
	n    int
	text string
}

// verifyFieldSep splits a field line that joins several fields:
// "Institute for Imaginary Epidemiology | Leiden, Netherlands".
const verifyFieldSep = "|·•"

// checkEntry traces a "###" block. The title must equal a history entry or
// work title, and every field line (or "|"-separated part of one) must
// equal a field of that same entry: organisation, location or date range
// for an entry; venue, authors or date for a work. With several entries
// of that title, the one that explains most fields is reported against.
// The title and each field part count as one claim.
func (c *verifyCorpus) checkEntry(title verifyLine, fields []verifyLine) (int, []Claim) {
	parts := fieldParts(fields)
	var best []Claim
	found := false
	for _, r := range c.records {
		if r.title != foldField(title.text) {
			continue
		}
		bad := r.unmatched(parts)
		if !found || len(bad) < len(best) {
			best, found = bad, true
		}
	}
	if !found {
		best = []Claim{{Line: title.n, Kind: KindEntry, Text: strings.TrimSpace(title.text)}}
	}
	return 1 + len(parts), best
}

func fieldParts(fields []verifyLine) []verifyLine {
	var parts []verifyLine
	for _, f := range fields {
		for _, p := range strings.FieldsFunc(f.text, func(c rune) bool {
			return strings.ContainsRune(verifyFieldSep, c)
		}) {
			if p = strings.TrimSpace(p); p != "" {
				parts = append(parts, verifyLine{n: f.n, text: p})
			}
		}
	}
	return parts
}

func (r verifyRecord) unmatched(parts []verifyLine) []Claim {
	var bad []Claim
	for _, p := range parts {
		if !r.matches(p.text) {
			bad = append(bad, Claim{Line: p.n, Kind: KindEntry, Text: p.text})
		}
	}
	return bad
}

// matches compares one field with the record, dates and authors by
// meaning: "June 2022" equals "Jun 2022", and "A; B" equals "A and B".
func (r verifyRecord) matches(field string) bool {
	if r.fields[foldField(field)] {
		return true
	}
	plain := strings.Trim(field, "*_`\\ ")
	if r.work {
		d, err := profile.ParseDate(plain)
		return (err == nil && d == r.start) || authorsKey(profile.SplitAuthors(plain)) == r.authors
	}
	s, e, err := profile.ParseRange(plain)
	return err == nil && s == r.start && e == r.end
}
