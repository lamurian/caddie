package document

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lamurian/caddie/internal/profile"
)

// verifyDate is a date found in text, normalised to "YYYY-MM-DD",
// "YYYY-MM" or "YYYY" as profile.ParseDate does.
type verifyDate struct{ norm, text string }

const verifyMonth = `(jan(?:uary)?|feb(?:ruary)?|mar(?:ch)?|apr(?:il)?|may|june?|july?|` +
	`aug(?:ust)?|sep(?:t(?:ember)?)?|oct(?:ober)?|nov(?:ember)?|dec(?:ember)?)\.?`

// verifyDatePatterns run in order; each match is blanked before the next
// pattern runs, so "1 June 2023" is one day date, not a month and a year.
var verifyDatePatterns = []struct {
	re   *regexp.Regexp
	norm func(m []string) (string, bool)
}{
	{regexp.MustCompile(`\b(\d{4})-(\d{2})(?:-(\d{2}))?\b`), isoDate},
	{regexp.MustCompile(`(?i)\b(\d{1,2})(?:st|nd|rd|th)?\s+` + verifyMonth + `,?\s+(\d{4})\b`),
		func(m []string) (string, bool) { return dayDate(m[3], m[2], m[1]) }},
	{regexp.MustCompile(`(?i)\b` + verifyMonth + `\s+(\d{1,2})(?:st|nd|rd|th)?,\s+(\d{4})\b`),
		func(m []string) (string, bool) { return dayDate(m[3], m[1], m[2]) }},
	{regexp.MustCompile(`(?i)\b` + verifyMonth + `,?\s+(\d{4})\b`),
		func(m []string) (string, bool) { return monthDate(m[2], m[1]) }},
	{regexp.MustCompile(`\b(?:19|20)\d{2}\b`),
		func(m []string) (string, bool) { return m[0], true }},
}

// scanVerifyDates finds every date in s and returns s with each one
// replaced by a comma, so the words around it do not join into a phrase.
func scanVerifyDates(s string) ([]verifyDate, string) {
	var out []verifyDate
	for _, p := range verifyDatePatterns {
		s = p.re.ReplaceAllStringFunc(s, func(match string) string {
			n, ok := p.norm(p.re.FindStringSubmatch(match))
			if !ok {
				return match
			}
			out = append(out, verifyDate{norm: n, text: strings.TrimSpace(match)})
			return " , "
		})
	}
	return out, s
}

func monthDate(year, month string) (string, bool) {
	n, err := profile.ParseDate(month + " " + year)
	return n, err == nil
}

func dayDate(year, month, day string) (string, bool) {
	ym, ok := monthDate(year, month)
	d, _ := strconv.Atoi(day)
	if !ok || d < 1 || d > 31 {
		return "", false
	}
	return fmt.Sprintf("%s-%02d", ym, d), true
}

// isoDate reads "YYYY-MM" or "YYYY-MM-DD", as answers.yml stores dates.
func isoDate(m []string) (string, bool) {
	mo, _ := strconv.Atoi(m[2])
	if mo < 1 || mo > 12 {
		return "", false
	}
	if m[3] == "" {
		return m[1] + "-" + m[2], true
	}
	d, _ := strconv.Atoi(m[3])
	if d < 1 || d > 31 {
		return "", false
	}
	return m[0], true
}

// addVerifyDate records n and the coarser dates it implies: "2027-01-01"
// also vouches for "2027-01" and "2027".
func addVerifyDate(set map[string]bool, n string) {
	for _, l := range []int{10, 7, 4} {
		if len(n) >= l {
			set[n[:l]] = true
		}
	}
}
