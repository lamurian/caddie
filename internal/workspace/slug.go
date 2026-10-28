package workspace

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// MaxSlugLen caps slug length so file names stay manageable.
const MaxSlugLen = 80

// Slug lowercases s, strips diacritics and joins ASCII letter and digit
// runs with single hyphens: "Société Générale" → "societe-generale".
func Slug(s string) string {
	folded, _, err := transform.String(transform.Chain(norm.NFD,
		runes.Remove(runes.In(unicode.Mn)), norm.NFC), s)
	if err != nil {
		folded = s
	}
	var b strings.Builder
	hyphen := false
	for _, r := range strings.ToLower(folded) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			if hyphen && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			hyphen = false
			continue
		}
		hyphen = true
	}
	return truncateSlug(b.String())
}

func truncateSlug(s string) string {
	if len(s) <= MaxSlugLen {
		return s
	}
	s = s[:MaxSlugLen]
	if i := strings.LastIndexByte(s, '-'); i > 0 {
		s = s[:i]
	}
	return s
}

// PostingSlug names a posting: "<title> at <company>" as a slug.
func PostingSlug(title, company string) string {
	return Slug(title + " at " + company)
}
