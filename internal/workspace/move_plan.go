package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// fileMove is one artefact to relocate.
type fileMove struct{ from, to string }

// movePlan lists the artefacts that exist (moves) and the ones that do not
// (skipped, workspace-relative source paths). dir is <dest>/<slug>/.
type movePlan struct {
	dir     string
	moves   []fileMove
	skipped []string
}

// planArtefacts finds the resume, cover letter and trace files to move into
// dir. Nothing is touched. Only regular files count.
func planArtefacts(l Layout, o MoveOptions, dir string) (movePlan, error) {
	plan := movePlan{dir: dir, skipped: []string{}}
	resumes, letters := l.Dir(DirResumes), l.Dir(DirCoverLetters)
	cands := []fileMove{
		{filepath.Join(resumes, o.Slug+".md"), filepath.Join(dir, "resume.md")},
		{filepath.Join(letters, o.Slug+".md"), filepath.Join(dir, "cover-letter.md")},
	}
	if o.ResumePDF != "" {
		cands = append(cands, fileMove{filepath.Join(resumes, o.ResumePDF), filepath.Join(dir, o.ResumePDF)})
	}
	if o.LetterPDF != "" {
		cands = append(cands, fileMove{filepath.Join(letters, o.LetterPDF), filepath.Join(dir, o.LetterPDF)})
	}
	for _, c := range cands {
		ok, err := isRegular(c.from)
		if err != nil {
			return movePlan{}, err
		}
		if ok {
			plan.moves = append(plan.moves, c)
		} else {
			plan.skipped = append(plan.skipped, rel(l.Root, c.from))
		}
	}
	if o.TraceDir == "" {
		return plan, nil
	}
	trace, err := traceFiles(o.TraceDir, filepath.Join(dir, "trace"))
	if err != nil {
		return movePlan{}, err
	}
	if len(trace) == 0 {
		plan.skipped = append(plan.skipped, rel(l.Root, o.TraceDir))
	}
	plan.moves = append(plan.moves, trace...)
	return plan, nil
}

// isRegular reports whether path is a regular file; absent is false, nil.
func isRegular(path string) (bool, error) {
	fi, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !fi.Mode().IsRegular() {
		return false, fmt.Errorf("%w: %s is not a regular file", ErrBadMove, path)
	}
	return true, nil
}

// traceFiles lists the regular files directly in src as moves into dst.
// The directory itself stays (it is a bind-mount source); a missing src
// yields no moves.
func traceFiles(src, dst string) ([]fileMove, error) {
	entries, err := os.ReadDir(src)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []fileMove
	for _, e := range entries {
		if e.Type().IsRegular() {
			out = append(out, fileMove{filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())})
		}
	}
	return out, nil
}

// keepInHandoff are the tracked files of handoff/ (ADR 0024): never cleared.
var keepInHandoff = map[string]bool{".gitignore": true, "README.md": true}

// ClearHandoff deletes every non-directory entry of dir except its tracked
// .gitignore and README.md. Subdirectories are left alone. A missing dir is
// nothing to clear.
func ClearHandoff(dir string) error {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("clear handoff: %w", err)
	}
	var errs []error
	for _, e := range entries {
		if e.IsDir() || keepInHandoff[e.Name()] {
			continue
		}
		if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
			errs = append(errs, fmt.Errorf("clear handoff: %w", err))
		}
	}
	return errors.Join(errs...)
}
