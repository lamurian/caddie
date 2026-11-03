package posting

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lamurian/caddie/internal/workspace"
)

// ErrNoSteps means SetManualSteps got no steps at all; ErrEmptyStep means
// one step was blank. Both are caller mistakes: the CLI maps them to a
// usage error.
var (
	ErrNoSteps   = errors.New("at least one manual step is required")
	ErrEmptyStep = errors.New("manual step is empty")
)

const manualHeading = "## Manual steps"

// Step is one `- [ ]` / `- [x]` item of a posting's `## Manual steps`.
type Step struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// Checklist is the `## Manual steps` section after a merge: every item in
// file order and how many this call added.
type Checklist struct {
	Path  string `json:"path"`
	Steps []Step `json:"steps"`
	Added int    `json:"added"`
}

var itemRe = regexp.MustCompile(`^\s*[-*] \[([ xX])\] (.*)$`)

// SetManualSteps writes or updates the `## Manual steps` checklist of the
// posting file at path (ADR 0024). The frontmatter bytes are kept as they
// are; only the body changes, and the file is rewritten only if it does.
func SetManualSteps(path string, steps []string) (Checklist, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Checklist{}, err
	}
	_, body, err := workspace.ParsePosting(path, data)
	if err != nil {
		return Checklist{}, err
	}
	newBody, list, err := MergeManualSteps(body, steps)
	if err != nil {
		return Checklist{}, err
	}
	list.Path = path
	if bytes.Equal(newBody, body) {
		return list, nil
	}
	front := data[:len(data)-len(body)]
	out := append(append([]byte{}, front...), newBody...)
	return list, writeFileAtomic(path, out)
}

// MergeManualSteps adds steps to the `## Manual steps` section of a posting
// body, creating the section at the end when missing. Existing items keep
// their order, text and tick; a step already present (trimmed,
// whitespace-collapsed, case-insensitive) is not added again.
func MergeManualSteps(body []byte, steps []string) ([]byte, Checklist, error) {
	clean, err := normalizeSteps(steps)
	if err != nil {
		return nil, Checklist{}, err
	}
	eol := "\n"
	if bytes.Contains(body, []byte("\r\n")) {
		eol = "\r\n"
	}
	lines := strings.SplitAfter(string(body), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	start, end := findSection(lines)
	if start < 0 {
		list := Checklist{Steps: newSteps(clean), Added: len(clean)}
		return appendSection(body, clean, eol), list, nil
	}
	out, list := mergeSection(lines, start, end, clean, eol)
	return out, list, nil
}

// normalizeSteps trims and collapses each step to one line and drops
// repeats within the input.
func normalizeSteps(steps []string) ([]string, error) {
	if len(steps) == 0 {
		return nil, ErrNoSteps
	}
	seen := map[string]bool{}
	var out []string
	for i, s := range steps {
		s = strings.Join(strings.Fields(s), " ")
		if s == "" {
			return nil, fmt.Errorf("%w: step %d", ErrEmptyStep, i+1)
		}
		if !seen[stepKey(s)] {
			seen[stepKey(s)] = true
			out = append(out, s)
		}
	}
	return out, nil
}

func stepKey(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }

func trimEOL(line string) string { return strings.TrimRight(line, "\r\n") }

// findSection returns the line range [start, end) of `## Manual steps`,
// ending at the next level-1 or level-2 heading; start is -1 when absent.
func findSection(lines []string) (start, end int) {
	start = -1
	for i, l := range lines {
		t := strings.TrimRight(trimEOL(l), " \t")
		if start < 0 {
			if t == manualHeading {
				start = i
			}
			continue
		}
		if strings.HasPrefix(t, "## ") || strings.HasPrefix(t, "# ") {
			return start, i
		}
	}
	return start, len(lines)
}

func newSteps(texts []string) []Step {
	out := make([]Step, len(texts))
	for i, t := range texts {
		out[i] = Step{Text: t}
	}
	return out
}

func itemLines(texts []string, eol string) string {
	var b strings.Builder
	for _, t := range texts {
		b.WriteString("- [ ] " + t + eol)
	}
	return b.String()
}

// appendSection adds a fresh section after the body, one blank line apart.
func appendSection(body []byte, steps []string, eol string) []byte {
	head := strings.TrimRight(string(body), " \t\r\n")
	if head != "" {
		head += eol + eol
	}
	return []byte(head + manualHeading + eol + eol + itemLines(steps, eol))
}

// mergeSection inserts the steps not yet listed after the section's last
// item, or, with no items yet, after its last non-blank line.
func mergeSection(lines []string, start, end int, steps []string, eol string) ([]byte, Checklist) {
	var list Checklist
	have := map[string]bool{}
	last := -1
	for i := start + 1; i < end; i++ {
		m := itemRe.FindStringSubmatch(trimEOL(lines[i]))
		if m == nil {
			continue
		}
		text := strings.TrimSpace(m[2])
		list.Steps = append(list.Steps, Step{Text: text, Done: m[1] != " "})
		have[stepKey(text)] = true
		last = i
	}
	var add []string
	for _, s := range steps {
		if !have[stepKey(s)] {
			add = append(add, s)
		}
	}
	list.Steps = append(list.Steps, newSteps(add)...)
	list.Added = len(add)
	if len(add) == 0 {
		return []byte(strings.Join(lines, "")), list
	}
	ins := itemLines(add, eol)
	if last < 0 {
		var blank bool
		if last, blank = firstItemSpot(lines, start, end); blank {
			ins = eol + ins
		}
	}
	if !strings.HasSuffix(lines[last], "\n") {
		lines[last] += eol
	}
	out := strings.Join(lines[:last+1], "") + ins + strings.Join(lines[last+1:], "")
	return []byte(out), list
}

// firstItemSpot is the line the first item goes after in a section with no
// items: its last non-blank line, or the blank line right after a bare
// heading. blank says whether a blank line must separate them.
func firstItemSpot(lines []string, start, end int) (at int, blank bool) {
	for i := end - 1; i > start; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return i, true
		}
	}
	if start+1 < end {
		return start + 1, false
	}
	return start, true
}

// writeFileAtomic replaces path through a temp file in the same directory,
// keeping its mode.
func writeFileAtomic(path string, content []byte) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_, err = tmp.Write(content)
	if err == nil {
		err = tmp.Chmod(fi.Mode().Perm())
	}
	if err == nil {
		err = tmp.Sync()
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	if err != nil {
		_ = os.Remove(name)
	}
	return err
}
