package posting

import (
	"strings"

	"github.com/lamurian/caddie/internal/source"
)

// Section indexes, in the order they appear in postings/<slug>.md.
const (
	Description = iota
	Responsibilities
	Requirements
	Perks
	numSections
)

// SectionNames are the `##` headings of PRD §12.3.
var SectionNames = [numSections]string{"Description", "Responsibilities", "Requirements", "Perks and benefits"}

// Sections holds the blocks of a description, sorted into the four sections.
type Sections [numSections][]source.Block

// cues map heading text to a section. They are checked in this order, so
// "What we offer" is a perk before "you will" makes it a responsibility.
var cues = []struct {
	section int
	words   []string
}{
	{Perks, []string{"perk", "benefit", "we offer", "what we give", "compensation", "why ",
		"what's in it for you", "package", "rewards"}},
	{Requirements, []string{"requirement", "qualification", "bring", "you have", "who you are",
		"about you", "your profile", "the profile", "candidate profile", "skills", "experience", "must have",
		"nice to have", "looking for", "ideal candidate", "competenc", "education"}},
	{Responsibilities, []string{"responsibilit", "what you'll do", "what you will do", "you will", "your role", "in this role", "duties", "tasks", "day to day", "day-to-day",
		"your impact", "your mission", "key accountabilities", "what you'll be doing"}},
}

// classify returns the section a heading starts, if it names one.
func classify(heading string) (int, bool) {
	h := strings.ReplaceAll(strings.ToLower(heading), "’", "'")
	for _, c := range cues {
		for _, w := range c.words {
			if strings.Contains(h, w) {
				return c.section, true
			}
		}
	}
	return Description, false
}

// Extract sorts description HTML into sections (ADR 0017). Text before the
// first recognised heading, and after any heading no cue matches, belongs
// to Description. A heading that repeats a job title is dropped.
func Extract(descHTML string, titles ...string) Sections {
	var s Sections
	cur := Description
	for _, b := range source.Blocks(descHTML) {
		if b.Kind != source.Heading {
			s[cur] = append(s[cur], b)
			continue
		}
		if isTitle(b.Text, titles) {
			continue
		}
		cur, _ = classify(b.Text)
		if !strings.EqualFold(b.Text, SectionNames[cur]) {
			s[cur] = append(s[cur], b)
		}
	}
	return s
}

func isTitle(heading string, titles []string) bool {
	for _, t := range titles {
		if strings.EqualFold(heading, t) {
			return true
		}
	}
	return false
}

// Missing lists the sections with no text.
func (s Sections) Missing() []string {
	out := []string{}
	for i, blocks := range s {
		if len(blocks) == 0 {
			out = append(out, SectionNames[i])
		}
	}
	return out
}

// notStated fills a section the posting has no text for.
const notStated = "_Not stated in the posting._"

// Markdown renders the four sections. Original sub-headings become bold
// lines; list items stay list items.
func (s Sections) Markdown() string {
	var b strings.Builder
	for i, blocks := range s {
		b.WriteString("## " + SectionNames[i] + "\n\n")
		if len(blocks) == 0 {
			b.WriteString(notStated + "\n\n")
			continue
		}
		b.WriteString(renderBlocks(blocks))
	}
	return b.String()
}

func renderBlocks(blocks []source.Block) string {
	var b strings.Builder
	for i, bl := range blocks {
		next := source.Para
		if i+1 < len(blocks) {
			next = blocks[i+1].Kind
		}
		switch bl.Kind {
		case source.Heading:
			b.WriteString("**" + bl.Text + "**\n\n")
		case source.Item:
			b.WriteString("- " + escape(bl.Text) + "\n")
			if next != source.Item {
				b.WriteString("\n")
			}
		default:
			b.WriteString(escape(bl.Text) + "\n\n")
		}
	}
	return b.String()
}

// escape keeps page text from turning into Markdown headings or fences.
func escape(s string) string {
	if strings.HasPrefix(s, "#") || strings.HasPrefix(s, "---") {
		return `\` + s
	}
	return s
}
