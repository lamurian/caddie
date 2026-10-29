package document

import "strings"

// proseCheck collects the claims of one line.
type proseCheck struct {
	c      *verifyCorpus
	n      int
	claims int
	bad    []Claim
	run    []verifyWord // adjacent term words, checked as one phrase
}

func (p *proseCheck) report(kind, text string, ok bool) {
	p.claims++
	if !ok {
		p.bad = append(p.bad, Claim{Line: p.n, Kind: kind, Text: text})
	}
}

func (p *proseCheck) flush() {
	if len(p.run) > 0 {
		p.report(KindTerm, joinRaw(p.run), p.c.hasTerm(p.run))
		p.run = nil
	}
}

// checkProse extracts the claims of one line of prose and returns those
// that do not trace. A line with no claim tokens is non-factual and
// passes. initial says whether the line's first word starts a sentence
// (false for frontmatter values, where every capital is a claim).
func (c *verifyCorpus) checkProse(n int, text string, initial bool) (claims int, bad []Claim) {
	p := &proseCheck{c: c, n: n}
	rest := stripURLs(text, func(u string) { p.report(KindTerm, u, c.urls[normURL(u)]) })
	dates, rest := scanVerifyDates(rest)
	for _, d := range dates {
		p.report(KindDate, d.text, c.dates[d.norm])
	}
	for _, w := range splitVerifyWords(rest) {
		p.word(w, initial)
		if w.key != "" || w.end {
			initial = w.end && !verifyAbbrev[w.key]
		}
	}
	p.flush()
	return p.claims, p.bad
}

// word adds w to the current term run, checks it as a number, or ends
// the run. Punctuation between two words also ends the run.
func (p *proseCheck) word(w verifyWord, initial bool) {
	if len(p.run) > 0 && (p.run[len(p.run)-1].stop || w.open) {
		p.flush()
	}
	if num, ok := verifyNumber(w); ok {
		p.flush()
		p.report(KindNumber, w.raw, p.c.numbers[num])
		return
	}
	if isTermShaped(w, initial) {
		p.run = append(p.run, w)
		return
	}
	p.flush()
}

// hasTerm checks a run of adjacent term words ("Fjordline Health
// Research") as one phrase, so known words cannot be recombined into an
// unknown name. A lone hyphenated word ("Utrecht-based") falls back to
// its term-shaped parts.
func (c *verifyCorpus) hasTerm(run []verifyWord) bool {
	keys := make([]string, len(run))
	for i, w := range run {
		keys[i] = w.key
	}
	if c.hasPhrase(strings.Join(keys, " "), true) {
		return true
	}
	if len(run) != 1 || !strings.Contains(run[0].raw, "-") {
		return false
	}
	for _, part := range strings.Split(run[0].raw, "-") {
		w := verifyWord{raw: part, key: foldClaim(part)}
		if isTermShaped(w, false) && !c.hasPhrase(w.key, true) {
			return false
		}
	}
	return true
}

func joinRaw(run []verifyWord) string {
	parts := make([]string, len(run))
	for i, w := range run {
		parts[i] = w.raw
	}
	return strings.Join(parts, " ")
}

// checkSkills traces one line of a skills section. Text before a colon is
// a label, checked as prose; the items after it, split on commas,
// semicolons, bullets and " and ", must each appear as a phrase in the
// profile or answers. Targets do not vouch for skills.
func (c *verifyCorpus) checkSkills(n int, text string) (claims int, bad []Claim) {
	if label, items, ok := strings.Cut(text, ":"); ok {
		claims, bad = c.checkProse(n, label, true)
		text = items
	}
	text = strings.ReplaceAll(text, " and ", ",")
	for _, item := range strings.FieldsFunc(text, func(r rune) bool { return strings.ContainsRune(",;•·|", r) }) {
		item = strings.Trim(strings.TrimSpace(item), "*_`.")
		phrase := strings.TrimSpace(verifyPhrases(item)[0])
		if phrase == "" {
			continue
		}
		claims++
		if !c.hasPhrase(phrase, false) {
			bad = append(bad, Claim{Line: n, Kind: KindSkill, Text: item})
		}
	}
	return claims, bad
}
