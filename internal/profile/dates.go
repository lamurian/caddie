package profile

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	monthNames = []string{"january", "february", "march", "april", "may", "june",
		"july", "august", "september", "october", "november", "december"}
	rangeSep = regexp.MustCompile(`^(.+?)\s*[–—-]\s*(.+)$`)
	yearRe   = regexp.MustCompile(`^\d{4}$`)
)

// date is a parsed "Mon YYYY" or "YYYY"; month is 0 when absent.
type date struct{ year, month int }

func (d date) String() string {
	if d.month == 0 {
		return strconv.Itoa(d.year)
	}
	return fmt.Sprintf("%04d-%02d", d.year, d.month)
}

func parseMonth(word string) (int, bool) {
	w := strings.ToLower(strings.TrimSuffix(word, "."))
	for i, name := range monthNames {
		if w == name || w == name[:3] || (w == "sept" && i == 8) {
			return i + 1, true
		}
	}
	return 0, false
}

// parseDate accepts "Aug 2022", "August 2022" or "2022".
func parseDate(s string) (date, error) {
	f := strings.Fields(s)
	if len(f) == 0 || len(f) > 2 || !yearRe.MatchString(f[len(f)-1]) {
		return date{}, fmt.Errorf("bad date %q, want \"Mon YYYY\" or \"YYYY\"", s)
	}
	year, _ := strconv.Atoi(f[len(f)-1])
	if len(f) == 1 {
		return date{year: year}, nil
	}
	month, ok := parseMonth(f[0])
	if !ok {
		return date{}, fmt.Errorf("bad month %q in %q", f[0], s)
	}
	return date{year: year, month: month}, nil
}

// ParseDate normalises "Aug 2022", "August 2022", "Sept 2014" or "2022" to
// "YYYY-MM" or "YYYY", the form used in Entry and Work. The document
// package uses it to compare dates in drafts with the history.
func ParseDate(s string) (string, error) {
	d, err := parseDate(s)
	if err != nil {
		return "", err
	}
	return d.String(), nil
}

// ParseRange accepts "Mon YYYY – Mon YYYY" or "Mon YYYY – Present" (en
// dash, em dash or hyphen) and returns normalised start and end strings.
func ParseRange(s string) (start, end string, err error) {
	m := rangeSep.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return "", "", fmt.Errorf("bad date range %q, want \"Mon YYYY – Mon YYYY|Present\"", s)
	}
	from, err := parseDate(m[1])
	if err != nil {
		return "", "", err
	}
	if strings.EqualFold(m[2], "present") {
		return from.String(), "present", nil
	}
	to, err := parseDate(m[2])
	if err != nil {
		return "", "", err
	}
	if to.year < from.year || (to.year == from.year && to.month != 0 && to.month < from.month) {
		return "", "", fmt.Errorf("date range %q ends before it starts", s)
	}
	return from.String(), to.String(), nil
}
