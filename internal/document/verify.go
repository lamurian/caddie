package document

import (
	"regexp"
	"sort"
	"strings"

	"github.com/lamurian/caddie/internal/profile"
)

// Claim kinds (ADR 0021).
const (
	KindEntry  = "entry"  // a "###" block: title, organisation, location, dates
	KindDate   = "date"   // "Jun 2023", "June 2023", "2019", "1 January 2027"
	KindNumber = "number" // "40,000", "60", "30%", "three"
	KindTerm   = "term"   // names, technical terms, URLs, e-mail addresses
	KindSkill  = "skill"  // an item under a skills heading
)

// Sources are what a draft may claim: the profile, the answer values in
// scope for the posting (global, company and posting), and the targets,
// the posting's company and title, which a letter may name.
type Sources struct {
	Profile profile.Profile
	Answers []string
	Targets []string
}

// Claim is one factual claim of a draft. Line is 1-based.
type Claim struct {
	Line int    `json:"line"`
	Kind string `json:"kind"`
	Text string `json:"text"`
}

// Report is the result of Verify: how many claims were checked and the
// ones that do not trace, in line order (never nil).
type Report struct {
	Claims      int     `json:"claims"`
	Untraceable []Claim `json:"untraceable"`
}

// OK reports whether every claim traced.
func (r Report) OK() bool { return len(r.Untraceable) == 0 }

var (
	verifyHeadingRe = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	verifyMarkerRe  = regexp.MustCompile(`^(?:>\s*)*(?:[-*+]|\d+[.)])\s+`)
	verifyKeyRe     = regexp.MustCompile(`^[\w-]+:(?:\s+(.*))?$`)
	verifySkillsRe  = regexp.MustCompile(`(?i)\b(skills?|tools|technologies|competencies|expertise)\b`)
)

type verifier struct {
	c      *verifyCorpus
	lines  []string
	skills bool
	inside bool // the previous prose line ended mid-sentence
	report Report
}

// Verify extracts the factual claims of a Markdown draft and traces each
// one to src (ADR 0004, ADR 0021). It never fails: a draft it cannot
// read as structure is checked as prose, which is stricter.
func Verify(doc []byte, src Sources) Report {
	text := strings.ReplaceAll(string(doc), "\r\n", "\n")
	v := &verifier{c: newVerifyCorpus(src), lines: strings.Split(text, "\n"),
		report: Report{Untraceable: []Claim{}}}
	for i := v.frontmatter(); i < len(v.lines); {
		i = v.line(i)
	}
	sort.SliceStable(v.report.Untraceable, func(a, b int) bool {
		return v.report.Untraceable[a].Line < v.report.Untraceable[b].Line
	})
	return v.report
}

func (v *verifier) add(claims int, bad []Claim) {
	v.report.Claims += claims
	v.report.Untraceable = append(v.report.Untraceable, bad...)
}

// frontmatter checks the values of a leading "---" block, every capital
// counting as a claim; keys are not claims. It returns the first body line.
func (v *verifier) frontmatter() int {
	if len(v.lines) == 0 || strings.TrimSpace(v.lines[0]) != "---" {
		return 0
	}
	end := 0
	for i := 1; i < len(v.lines) && end == 0; i++ {
		if t := strings.TrimSpace(v.lines[i]); t == "---" || t == "..." {
			end = i
		}
	}
	for i := 1; i < end; i++ {
		t := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v.lines[i]), "- "))
		if m := verifyKeyRe.FindStringSubmatch(t); m != nil {
			t = m[1]
		}
		if t != "" && !strings.HasPrefix(t, "#") && t != "|" && t != ">" {
			v.add(v.c.checkProse(i+1, t, false))
		}
	}
	if end == 0 {
		return 0
	}
	return end + 1
}

// line checks line i and returns the next line to check. "#" and "##"
// headings are structure; "##" headings also open or close a skills
// section. "###" starts an entry block; deeper headings are prose.
func (v *verifier) line(i int) int {
	t := strings.TrimSpace(v.lines[i])
	m := verifyHeadingRe.FindStringSubmatch(t)
	inside := v.inside
	v.inside = false
	switch {
	case t == "":
	case m != nil && len(m[1]) <= 2:
		v.skills = verifySkillsRe.MatchString(m[2])
	case m != nil && len(m[1]) == 3:
		return v.entry(i, m[2])
	case m != nil:
		v.add(v.c.checkProse(i+1, m[2], true))
	case v.skills:
		v.add(v.c.checkSkills(i+1, verifyMarkerRe.ReplaceAllString(t, "")))
	default:
		v.prose(i, t, inside)
	}
	return i + 1
}

// prose checks a paragraph or list line. Its first word starts a sentence
// only after a list marker, a blank line, a heading or a line that ended
// one, so a capital that opens a wrapped line is still a claim.
func (v *verifier) prose(i int, t string, inside bool) {
	marker := verifyMarkerRe.FindString(t)
	body := t[len(marker):]
	v.add(v.c.checkProse(i+1, body, marker != "" || !inside))
	end := strings.TrimRight(body, "*_\\\"'”’) ")
	v.inside = !strings.HasSuffix(end, ".") && !strings.HasSuffix(end, "!") &&
		!strings.HasSuffix(end, "?") && !strings.HasSuffix(end, "…")
}

// entry checks a "###" title and the field lines that directly follow it
// (at most three, up to a blank line, list item or heading), as one claim.
func (v *verifier) entry(i int, title string) int {
	var fields []verifyLine
	j := i + 1
	for ; j < len(v.lines) && len(fields) < 3; j++ {
		t := strings.TrimSpace(v.lines[j])
		if t == "" || strings.HasPrefix(t, "#") || verifyMarkerRe.MatchString(t) {
			break
		}
		fields = append(fields, verifyLine{n: j + 1, text: t})
	}
	v.add(v.c.checkEntry(verifyLine{n: i + 1, text: title}, fields))
	return j
}
