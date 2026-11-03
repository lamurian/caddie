package document

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/lamurian/caddie/internal/workspace"
)

// maxNameLen caps the <Name> part so that the whole PDF name stays under
// the usual 255-byte file name limit next to two capped slugs.
const maxNameLen = 64

// PDFName is the rendered file name (PRD §12.6, ADR 0022):
// <Name>-<kind>-<position-slug>-<company-slug>.pdf, where <Name> is name
// with diacritics folded, case kept and every run of other characters
// replaced by one hyphen ("José Ñúñez O'Brien" → "Jose-Nunez-O-Brien").
// Empty parts are left out rather than leaving stray hyphens.
func PDFName(name string, kind Kind, position, company string) string {
	parts := make([]string, 0, 4)
	for _, p := range []string{nameSlug(name), string(kind), workspace.Slug(position), workspace.Slug(company)} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "-") + ".pdf"
}

// nameSlug folds diacritics, keeps case and joins ASCII letter and digit
// runs with single hyphens, capped at maxNameLen on a hyphen.
func nameSlug(s string) string {
	folded, _, err := transform.String(transform.Chain(norm.NFD,
		runes.Remove(runes.In(unicode.Mn)), norm.NFC), s)
	if err != nil {
		folded = s
	}
	var b strings.Builder
	hyphen := false
	for _, r := range folded {
		if r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			if hyphen && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			hyphen = false
			continue
		}
		hyphen = true
	}
	out := b.String()
	if len(out) <= maxNameLen {
		return out
	}
	out = out[:maxNameLen]
	if i := strings.LastIndexByte(out, '-'); i > 0 {
		out = out[:i]
	}
	return out
}
