package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lamurian/caddie/internal/posting"
	"github.com/lamurian/caddie/internal/workspace"
)

// postingSteps writes or merges the `## Manual steps` checklist of a posting
// in postings/ or manual/ (ADR 0026). An applied posting is refused (exit 6).
func postingSteps(args []string, e env) error {
	fs := newFlags("posting steps", e)
	fs.Bool("json", true, "print JSON (always on)")
	pos, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return usageErr("want `posting steps <slug> \"<step>\"…`")
	}
	slug := pos[0]
	if err := checkSlug(slug); err != nil {
		return err
	}
	_, layout, err := loadLayout(e)
	if err != nil {
		return err
	}
	path, err := stepsPosting(layout, slug)
	if err != nil {
		return err
	}
	list, err := posting.SetManualSteps(path, pos[1:])
	if err != nil {
		return err
	}
	list.Path = wsPath(layout, path)
	return writeJSON(e.stdout, list)
}

// stepsPosting finds <slug>.md in postings/, else manual/; one in applied/
// is an illegal transition, one nowhere is not found.
func stepsPosting(layout workspace.Layout, slug string) (string, error) {
	for _, dir := range []string{workspace.DirPostings, workspace.DirManual} {
		path := filepath.Join(layout.Dir(dir), slug+".md")
		_, err := os.Stat(path)
		if err == nil {
			return path, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	if _, err := os.Stat(filepath.Join(layout.Dir(workspace.DirApplied), slug+".md")); err == nil {
		return "", fmt.Errorf("%w: %s is already in applied/", workspace.ErrIllegalTransition, slug)
	}
	return "", fmt.Errorf("%w: no posting %q in postings/ or manual/", errNotFound, slug)
}
