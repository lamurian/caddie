package document

import (
	"fmt"
	"strings"

	"github.com/lamurian/caddie/internal/profile"
)

// Keyword statuses. A keyword is covered only where the history supports
// it (ADR 0004), so a draft hit without history support is flagged.
const (
	KeywordCovered     = "covered"     // in the draft and in history/
	KeywordMissing     = "missing"     // in history/, not in the draft: the agent may add it
	KeywordUnsupported = "unsupported" // not in history/: never add it
	KeywordUnbacked    = "unbacked"    // in the draft, not in history/: verify and remove it
)

// keywordOrder is the order Text groups keywords in.
var keywordOrder = []string{KeywordCovered, KeywordMissing, KeywordUnsupported, KeywordUnbacked}

// Keyword is one requirement phrase and where it was found.
type Keyword struct {
	Text      string `json:"keyword"`
	InDraft   bool   `json:"in_draft"`
	InHistory bool   `json:"in_history"`
	Status    string `json:"status"`
}

// KeywordReport is the advisory coverage report behind `caddie keywords`.
// Keywords keep the order they first appear in the posting.
type KeywordReport struct {
	Keywords []Keyword `json:"keywords"`
	Covered  int       `json:"covered"`
	Total    int       `json:"total"`
	Notices  []string  `json:"notices"`
}

// Keywords extracts keywords from the posting's "## Requirements" section
// and reports which the draft covers and whether the profile supports each.
// It never fails: a posting without requirements yields a notice instead.
func Keywords(posting, draft []byte, prof profile.Profile) KeywordReport {
	phrases, notices := requirementPhrases(posting)
	inDraft := markdownTokens(string(keywordsBody(draft)))
	inHistory := markdownTokens(historyText(prof))
	r := KeywordReport{Keywords: []Keyword{}, Notices: append([]string{}, notices...)}
	unbacked := false
	for _, p := range phrases {
		k := Keyword{Text: p.text, InDraft: containsRun(inDraft, p.tokens), InHistory: containsRun(inHistory, p.tokens)}
		k.Status = keywordStatus(k)
		if k.Status == KeywordCovered {
			r.Covered++
		}
		unbacked = unbacked || k.Status == KeywordUnbacked
		r.Keywords = append(r.Keywords, k)
	}
	r.Total = len(r.Keywords)
	if unbacked {
		r.Notices = append(r.Notices,
			"unbacked keywords are in the draft but not in history/: run caddie verify and remove them")
	}
	return r
}

func keywordStatus(k Keyword) string {
	switch {
	case k.InHistory && k.InDraft:
		return KeywordCovered
	case k.InHistory:
		return KeywordMissing
	case k.InDraft:
		return KeywordUnbacked
	default:
		return KeywordUnsupported
	}
}

// Text renders the report for the agent: a count line, one
// "<status> <keyword>" line per keyword grouped by status, then notes.
func (r KeywordReport) Text() string {
	counts := map[string]int{}
	for _, k := range r.Keywords {
		counts[k.Status]++
	}
	var b strings.Builder
	fmt.Fprintf(&b, "keywords: %d/%d covered, %d missing, %d unsupported, %d unbacked\n",
		r.Covered, r.Total, counts[KeywordMissing], counts[KeywordUnsupported], counts[KeywordUnbacked])
	for _, s := range keywordOrder {
		for _, k := range r.Keywords {
			if k.Status == s {
				fmt.Fprintf(&b, "%-12s %s\n", s, k.Text)
			}
		}
	}
	for _, n := range r.Notices {
		fmt.Fprintf(&b, "note: %s\n", n)
	}
	return b.String()
}

// keywordsBody drops a leading YAML frontmatter block ("---" … "---").
func keywordsBody(data []byte) []byte {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return data
	}
	lines := strings.SplitAfter(s, "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return []byte(strings.Join(lines[i+1:], ""))
		}
	}
	return data
}
