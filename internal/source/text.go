package source

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// SummaryLen caps Job.Summary, in runes, so search output stays short.
const SummaryLen = 600

func summarise(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if utf8.RuneCountInString(text) <= SummaryLen {
		return text
	}
	r := []rune(text)[:SummaryLen]
	return strings.TrimSpace(string(r)) + "…"
}

// WorkMode infers wfo|hybrid|wfh|wfa from free text such as a location or
// a description, or returns "" when the text does not say.
func WorkMode(texts ...string) string {
	t := strings.ToLower(strings.Join(texts, " "))
	remote := containsAny(t, "remote", "work from home")
	switch {
	case containsAny(t, "hybrid"):
		return "hybrid"
	case remote && containsAny(t, "anywhere", "worldwide"):
		return "wfa"
	case remote:
		return "wfh"
	case containsAny(t, "on-site", "onsite", "on site", "in office", "in-office"):
		return "wfo"
	}
	return ""
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// Salary is a parsed salary range; zero values mean unknown.
type Salary struct {
	Min, Max         int
	Currency, Period string
}

func (s Salary) apply(j *Job) {
	j.SalaryMin, j.SalaryMax, j.SalaryCurrency, j.SalaryPeriod = s.Min, s.Max, s.Currency, s.Period
}

var (
	salaryNum    = regexp.MustCompile(`(\d{1,3}(?:[,.\s]\d{3})+|\d+(?:\.\d+)?)\s*([kK])?`)
	currencyCode = regexp.MustCompile(`(?i)\b(usd|eur|gbp|chf|cad|aud)\b`)
	currencySym  = [][2]string{{"$", "USD"}, {"€", "EUR"}, {"£", "GBP"}}
)

// ParseSalary reads free text like "$80k - $100k per year" or
// "€60,000–75,000". It returns the zero Salary when no amount is found.
func ParseSalary(text string) Salary {
	var s Salary
	nums := salaryNum.FindAllStringSubmatch(text, 2)
	for i, m := range nums {
		v := parseAmount(m[1], m[2] != "")
		if v <= 0 {
			continue
		}
		if i == 0 || s.Min == 0 {
			s.Min = v
		}
		s.Max = v
	}
	if s.Min == 0 {
		return Salary{}
	}
	for _, c := range currencySym {
		if strings.Contains(text, c[0]) {
			s.Currency = c[1]
		}
	}
	if m := currencyCode.FindString(text); m != "" {
		s.Currency = strings.ToUpper(m)
	}
	s.Period = periodOf(strings.ToLower(text), s.Max)
	return s
}

func parseAmount(s string, thousands bool) int {
	clean := strings.NewReplacer(",", "", " ", "").Replace(s)
	if strings.Count(clean, ".") == 1 && len(clean)-strings.Index(clean, ".") == 4 {
		clean = strings.ReplaceAll(clean, ".", "") // 60.000 is sixty thousand
	}
	f, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0
	}
	if thousands {
		f *= 1000
	}
	return int(f)
}

// periodOf reads the pay period; an unstated period of 10,000 or more is
// taken as yearly.
func periodOf(lower string, amount int) string {
	switch {
	case strings.Contains(lower, "hour") || strings.Contains(lower, "/hr"):
		return "hour"
	case strings.Contains(lower, "month"):
		return "month"
	case strings.Contains(lower, "year") || strings.Contains(lower, "annual") ||
		strings.Contains(lower, "/yr") || amount >= 10000:
		return "year"
	}
	return ""
}

var deadlinePattern = regexp.MustCompile(`(?i)(?:application deadline|deadline|closing date|apply by|applications close)\s*(?:is|on|:)?\s*` +
	`(\d{4}-\d{2}-\d{2}|\d{1,2}\s+[A-Za-z]+\.?\s+\d{4}|[A-Za-z]+\.?\s+\d{1,2},?\s+\d{4})`)

var dateLayouts = []string{"2 January 2006", "2 Jan 2006", "January 2 2006", "Jan 2 2006"}

// Deadline finds an application deadline stated in text and returns it as
// YYYY-MM-DD, or "" when none is stated.
func Deadline(text string) string {
	m := deadlinePattern.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return Date(m[1])
}

// Date normalises a date in a common layout to YYYY-MM-DD, or returns "".
func Date(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 10 {
		if t, err := time.Parse("2006-01-02", s[:10]); err == nil {
			return t.Format("2006-01-02")
		}
	}
	s = strings.NewReplacer(",", "", ".", "").Replace(strings.TrimSpace(s))
	s = strings.Join(strings.Fields(s), " ")
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return ""
}

// unixDate formats seconds since the epoch as YYYY-MM-DD (UTC).
func unixDate(sec int64) string {
	if sec <= 0 {
		return ""
	}
	return time.Unix(sec, 0).UTC().Format("2006-01-02")
}
