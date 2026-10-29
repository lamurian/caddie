package profile

import (
	"strings"
)

type line struct {
	n    int
	text string
}

// block is one "### title" entry with its non-empty lines.
type block struct {
	title line
	lines []line
}

type sectionParser struct {
	file    string
	heading string
	blocks  []block
}

// parseSection splits a history file into its "##" heading and its
// "###" blocks, rejecting anything outside that structure.
func parseSection(file string, data []byte) (string, []block, error) {
	p := &sectionParser{file: file}
	for i, raw := range strings.Split(string(data), "\n") {
		text := strings.TrimRight(raw, " \t\r")
		if strings.TrimSpace(text) == "" {
			continue
		}
		if err := p.add(line{n: i + 1, text: text}); err != nil {
			return "", nil, err
		}
	}
	if p.heading == "" {
		return "", nil, errAt(file, 1, "missing \"## Section\" heading")
	}
	return p.heading, p.blocks, nil
}

func (p *sectionParser) add(l line) error {
	switch {
	case strings.HasPrefix(l.text, "### "):
		title := strings.TrimSpace(l.text[4:])
		p.blocks = append(p.blocks, block{title: line{n: l.n, text: title}})
	case strings.HasPrefix(l.text, "## "):
		return p.setHeading(l)
	case strings.HasPrefix(l.text, "#"):
		return errAt(p.file, l.n, "unexpected heading %q: use one ## section and ### entries", l.text)
	case len(p.blocks) == 0:
		return errAt(p.file, l.n, "text before the first ### entry: %q", l.text)
	default:
		last := &p.blocks[len(p.blocks)-1]
		last.lines = append(last.lines, l)
	}
	return nil
}

func (p *sectionParser) setHeading(l line) error {
	if p.heading != "" {
		return errAt(p.file, l.n, "second ## heading: each file has exactly one section")
	}
	if len(p.blocks) > 0 {
		return errAt(p.file, l.n, "## heading must come before the ### entries")
	}
	p.heading = strings.TrimSpace(l.text[3:])
	return nil
}

func isBullet(t string) bool {
	return strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ")
}

// splitBody separates the field lines that follow a title from its
// bullets. An indented line after a bullet continues that bullet.
func splitBody(file string, b block) (fields []line, bullets []string, err error) {
	for _, l := range b.lines {
		t := strings.TrimSpace(l.text)
		switch {
		case isBullet(t):
			bullets = append(bullets, strings.TrimSpace(t[2:]))
		case len(bullets) > 0 && t != l.text:
			bullets[len(bullets)-1] += " " + t
		case len(bullets) > 0:
			return nil, nil, errAt(file, l.n, "expected a bullet (- …) or the next ### entry, got %q", t)
		default:
			fields = append(fields, line{n: l.n, text: t})
		}
	}
	return fields, bullets, nil
}

// toEntry reads organisation, optional location and date range.
func toEntry(file string, b block) (Entry, error) {
	fields, bullets, err := splitBody(file, b)
	if err != nil {
		return Entry{}, err
	}
	if len(fields) < 2 || len(fields) > 3 {
		return Entry{}, errAt(file, b.title.n,
			"entry %q needs organisation, location (optional) and date range lines; found %d",
			b.title.text, len(fields))
	}
	dates := fields[len(fields)-1]
	start, end, err := ParseRange(dates.text)
	if err != nil {
		return Entry{}, errAt(file, dates.n, "%v", err)
	}
	e := Entry{Title: b.title.text, Organisation: fields[0].text, Start: start, End: end, Bullets: bullets}
	if len(fields) == 3 {
		e.Location = fields[1].text
	}
	return e, nil
}

// toWork reads venue, authors and date for a publication or presentation.
func toWork(file string, b block) (Work, error) {
	fields, bullets, err := splitBody(file, b)
	if err != nil {
		return Work{}, err
	}
	if len(fields) != 3 {
		return Work{}, errAt(file, b.title.n,
			"%q needs venue, authors and date lines; found %d", b.title.text, len(fields))
	}
	d, err := parseDate(fields[2].text)
	if err != nil {
		return Work{}, errAt(file, fields[2].n, "%v", err)
	}
	return Work{Title: b.title.text, Venue: fields[0].text, Authors: SplitAuthors(fields[1].text),
		Date: d.String(), Bullets: bullets}, nil
}

// SplitAuthors splits "A, B and C" or "A; B; C" into names.
func SplitAuthors(s string) []string {
	s = strings.ReplaceAll(s, ";", ",")
	s = strings.ReplaceAll(s, " and ", ", ")
	var out []string
	for _, a := range strings.Split(s, ",") {
		a = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(a), "and "))
		if a != "" {
			out = append(out, a)
		}
	}
	return out
}
