package source

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseSalary(t *testing.T) {
	cases := map[string]Salary{
		"$80k - $100k":               {80000, 100000, "USD", "year"},
		"€60,000–75,000 per year":    {60000, 75000, "EUR", "year"},
		"EUR 4.500 - 5.500 / month":  {4500, 5500, "EUR", "month"},
		"£40 per hour":               {40, 40, "GBP", "hour"},
		"£45 - £60 per hour":         {45, 60, "GBP", "hour"},
		"120000 CHF":                 {120000, 120000, "CHF", "year"},
		"Competitive":                {},
		"":                           {},
		"academic audit, 55000 cash": {55000, 55000, "", "year"},
	}
	for in, want := range cases {
		if got := ParseSalary(in); got != want {
			t.Errorf("ParseSalary(%q) = %+v, want %+v", in, got, want)
		}
	}
}

func TestDeadlineAndDate(t *testing.T) {
	cases := map[string]string{
		"Application deadline: 14 November 2026.": "2026-11-14",
		"Closing date: December 5, 2026":          "2026-12-05",
		"Apply by 2026-11-30 at noon":             "2026-11-30",
		"Applications close on 1 Dec 2026":        "2026-12-01",
		"Deadline is Sept 3 2026":                 "",
		"No deadline mentioned, 3 November 2026":  "",
	}
	for in, want := range cases {
		if got := Deadline(in); got != want {
			t.Errorf("Deadline(%q) = %q, want %q", in, got, want)
		}
	}
	if got := Date("2026-11-20T23:59:59.000+01:00"); got != "2026-11-20" {
		t.Errorf("Date(timestamp) = %q", got)
	}
}

func TestWorkMode(t *testing.T) {
	cases := map[string]string{
		"Amsterdam (Hybrid)":   "hybrid",
		"Remote - Worldwide":   "wfa",
		"Remote (EU)":          "wfh",
		"On-site in Berlin":    "wfo",
		"Utrecht, Netherlands": "",
	}
	for in, want := range cases {
		if got := WorkMode(in); got != want {
			t.Errorf("WorkMode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBlocks(t *testing.T) {
	in := "<p>Intro <b>bold</b> text.</p><h3>What you'll do</h3><ul><li><p>Build</p></li><li>Test</li></ul>" +
		"<p><strong>What you bring:</strong></p><p>Your profile:</p><div>Line one<br>Line two</div>"
	want := []Block{{Para, "Intro bold text."}, {Heading, "What you'll do"}, {Item, "Build"}, {Item, "Test"},
		{Heading, "What you bring"}, {Heading, "Your profile"}, {Para, "Line one"}, {Para, "Line two"}}
	if got := Blocks(in); !reflect.DeepEqual(got, want) {
		t.Errorf("Blocks:\n got %+v\nwant %+v", got, want)
	}
	if got := Blocks("&lt;p&gt;Escaped &amp;amp; fine&lt;/p&gt;"); len(got) != 1 || got[0].Text != "Escaped & fine" {
		t.Errorf("escaped HTML: %+v", got)
	}
}

func TestSummaryIsCapped(t *testing.T) {
	s := summarise(strings.Repeat("x", 1000))
	if n := len([]rune(s)); n != SummaryLen+1 || !strings.HasSuffix(s, "…") {
		t.Errorf("summary has %d runes", n)
	}
}

func TestRedact(t *testing.T) {
	if got := redact("https://u:p@api.example.org/x?app_key=secret"); got != "https://api.example.org/x" {
		t.Errorf("redact = %q", got)
	}
}
