package workspace

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// postingFile is one parsed <dir>/<slug>.md.
type postingFile struct {
	rel     string // workspace-relative, slash-separated
	posting Posting
	body    []byte
}

// readPostings parses every <dir>/*.md posting, in name order. README.md,
// dot files and subdirectories (the <slug>/ artefact directories) are
// skipped; a missing directory yields none.
func readPostings(l Layout, dir string) ([]postingFile, error) {
	path := l.Dir(dir)
	entries, err := os.ReadDir(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []postingFile
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") || name == "README.md" ||
			strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(path, name)
		data, err := os.ReadFile(full)
		if err != nil {
			return nil, err
		}
		r := rel(l.Root, full)
		p, body, err := ParsePosting(r, data)
		if err != nil {
			return nil, err
		}
		out = append(out, postingFile{rel: r, posting: p, body: body})
	}
	return out, nil
}

// manualHeading opens the checklist that `caddie posting steps` writes
// (ADR 0026); stepRe matches one of its items.
const manualHeading = "## Manual steps"

var stepRe = regexp.MustCompile(`^\s*[-*] \[([ xX])\] (.*)$`)

// openSteps returns the text of each unticked item in the body's
// `## Manual steps` section, in file order. The section ends at the next
// level-1 or level-2 heading, as internal/posting reads it; items
// elsewhere in the body are not counted.
func openSteps(body []byte) []string {
	open := []string{}
	in := false
	for _, line := range strings.Split(string(body), "\n") {
		t := strings.TrimRight(line, " \t\r")
		if !in {
			in = t == manualHeading
			continue
		}
		if strings.HasPrefix(t, "## ") || strings.HasPrefix(t, "# ") {
			break
		}
		if m := stepRe.FindStringSubmatch(t); m != nil && m[1] == " " {
			open = append(open, strings.TrimSpace(m[2]))
		}
	}
	return open
}
