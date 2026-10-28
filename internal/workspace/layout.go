package workspace

import (
	"os"
	"path/filepath"

	"github.com/lamurian/caddie/internal/config"
)

// Workspace directory names, relative to the workspace root (PRD §11).
const (
	DirPostings     = "postings"
	DirApplied      = "applied"
	DirManual       = "manual"
	DirResumes      = "resumes"
	DirCoverLetters = "cover-letters"
	DirReviews      = "reviews"
	DirHistory      = "history"
)

// Layout resolves workspace paths from the configuration.
type Layout struct {
	Root        string
	Handoff     string
	State       string
	Preferences string
	Answers     string
}

// New builds a Layout; relative config paths resolve against the workspace.
func New(cfg config.Config) Layout {
	return Layout{
		Root:        cfg.Workspace,
		Handoff:     cfg.Resolve(cfg.HandoffDir),
		State:       cfg.Resolve(cfg.StateDir),
		Preferences: cfg.Resolve(cfg.Preferences),
		Answers:     cfg.Resolve(cfg.Answers),
	}
}

// Dir returns the path of a named workspace directory.
func (l Layout) Dir(name string) string { return filepath.Join(l.Root, name) }

// History returns the path of the history directory.
func (l Layout) History() string { return l.Dir(DirHistory) }

// Required lists the directories that must exist in a healthy workspace.
// The state directory is not required: Caddie creates it on first write.
func (l Layout) Required() []string {
	dirs := []string{DirPostings, DirApplied, DirManual, DirResumes,
		DirCoverLetters, DirReviews, DirHistory}
	out := make([]string, 0, len(dirs)+1)
	for _, d := range dirs {
		out = append(out, l.Dir(d))
	}
	return append(out, l.Handoff)
}

// Missing returns the required directories that do not exist.
func (l Layout) Missing() []string {
	var missing []string
	for _, d := range l.Required() {
		if fi, err := os.Stat(d); err != nil || !fi.IsDir() {
			missing = append(missing, d)
		}
	}
	return missing
}
