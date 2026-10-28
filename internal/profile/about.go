package profile

import (
	"bytes"
	"strings"

	"go.yaml.in/yaml/v3"
)

// parseAbout reads the about.md frontmatter (identity) and its "## About"
// narrative. YAML errors carry line numbers that match the file.
func parseAbout(file string, data []byte) (Identity, string, error) {
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return Identity{}, "", errAt(file, 1, "about.md must start with --- frontmatter")
	}
	end := closingFence(lines)
	if end < 0 {
		return Identity{}, "", errAt(file, 1, "frontmatter has no closing ---")
	}
	var id Identity
	front := "\n" + strings.Join(lines[1:end], "\n") // keep file line numbers
	dec := yaml.NewDecoder(bytes.NewReader([]byte(front)))
	dec.KnownFields(true)
	if err := dec.Decode(&id); err != nil {
		return Identity{}, "", errAt(file, 0, "frontmatter: %v", err)
	}
	if strings.TrimSpace(id.Name) == "" {
		return Identity{}, "", errAt(file, 2, "frontmatter: name is required")
	}
	about, err := aboutBody(file, lines, end+1)
	return id, about, err
}

func closingFence(lines []string) int {
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return i
		}
	}
	return -1
}

// aboutBody returns the text after the "## About" heading.
func aboutBody(file string, lines []string, from int) (string, error) {
	for i := from; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if t == "" {
			continue
		}
		if t != "## About" {
			return "", errAt(file, i+1, "expected \"## About\" after the frontmatter, got %q", t)
		}
		about := strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
		if about == "" {
			return "", errAt(file, i+1, "the ## About section is empty")
		}
		return about, nil
	}
	return "", errAt(file, len(lines), "missing \"## About\" section")
}
