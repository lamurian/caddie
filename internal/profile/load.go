package profile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// workFiles use venue/authors/date instead of organisation/location/dates.
var workFiles = map[string]bool{"publication": true, "presentation": true}

// Load parses dir (normally history/): about.md is required; every other
// *.md file except README.md is a section, returned in file-name order.
func Load(dir string) (Profile, error) {
	var p Profile
	aboutPath := filepath.Join(dir, "about.md")
	data, err := os.ReadFile(aboutPath)
	if errors.Is(err, fs.ErrNotExist) {
		return p, errAt(aboutPath, 0, "missing: about.md is the identity source (see history/README.md)")
	}
	if err != nil {
		return p, err
	}
	if p.Identity, p.About, err = parseAbout(aboutPath, data); err != nil {
		return p, err
	}
	files, err := sectionFiles(dir)
	if err != nil {
		return p, err
	}
	for _, f := range files {
		s, err := loadSection(filepath.Join(dir, f))
		if err != nil {
			return p, err
		}
		p.Sections = append(p.Sections, s)
	}
	return p, nil
}

func sectionFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".md" || name == "about.md" || name == "README.md" {
			continue
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}

func loadSection(path string) (Section, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Section{}, err
	}
	heading, blocks, err := parseSection(path, data)
	if err != nil {
		return Section{}, err
	}
	s := Section{ID: strings.TrimSuffix(filepath.Base(path), ".md"), Heading: heading}
	for _, b := range blocks {
		if err := s.add(path, b); err != nil {
			return Section{}, err
		}
	}
	return s, nil
}

func (s *Section) add(path string, b block) error {
	if workFiles[s.ID] {
		w, err := toWork(path, b)
		s.Works = append(s.Works, w)
		return err
	}
	e, err := toEntry(path, b)
	s.Entries = append(s.Entries, e)
	return err
}
