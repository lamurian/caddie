package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Sentinel errors for Move (ADR 0026).
var (
	ErrDestExists = errors.New("destination already exists")
	ErrBadMove    = errors.New("invalid move options")
)

// MoveOptions says what `caddie move` archives and how.
type MoveOptions struct {
	Slug         string
	To           Status // StatusApplied or StatusManual
	Reason       string // manual_reason: required (one line) for manual, empty for applied
	Today        string // YYYY-MM-DD; required for applied
	FollowUpDays int    // apply.follow_up_days; ≥ 0
	ResumePDF    string // base name of the resume PDF (document.PDFName); "" = none
	LetterPDF    string // base name of the cover-letter PDF; "" = none
	TraceDir     string // host trace directory (<state>/trace); "" = none
}

// MoveResult is what `caddie move` prints. Paths are workspace-relative and
// slash-separated (absolute when outside the workspace).
type MoveResult struct {
	Slug          string   `json:"slug"`
	Status        Status   `json:"status"`
	Posting       string   `json:"posting"`
	Moved         []string `json:"moved"`
	Skipped       []string `json:"skipped"`
	AppliedAt     string   `json:"applied_at"`
	FollowUpAfter string   `json:"follow_up_after"`
	ManualReason  string   `json:"manual_reason"`
}

// Move relocates postings/<slug>.md to <dest>/<slug>.md and its artefacts
// into <dest>/<slug>/ (ADR 0026), updating the frontmatter. Every check
// runs before anything moves; a failure while moving is rolled back. After
// success handoff/ is cleared; a failure there is returned with the result
// and does not undo the move. Errors wrap ErrBadMove, ErrIllegalTransition,
// ErrDestExists, ErrBadPosting or os.ErrNotExist.
func Move(l Layout, o MoveOptions) (MoveResult, error) {
	if err := checkOptions(o); err != nil {
		return MoveResult{}, err
	}
	src := filepath.Join(l.Dir(DirPostings), o.Slug+".md")
	p, body, mode, err := readSource(l, src, o)
	if err != nil {
		return MoveResult{}, err
	}
	dest := l.Dir(string(o.To))
	if err := checkDest(dest, o.Slug); err != nil {
		return MoveResult{}, err
	}
	if err := setFrontmatter(&p, o); err != nil {
		return MoveResult{}, err
	}
	plan, err := planArtefacts(l, o, filepath.Join(dest, o.Slug))
	if err != nil {
		return MoveResult{}, err
	}
	post := filepath.Join(dest, o.Slug+".md")
	if err := execute(plan, src, post, append(p.Frontmatter(), body...), mode); err != nil {
		return MoveResult{}, err
	}
	res := MoveResult{
		Slug: o.Slug, Status: o.To, Posting: rel(l.Root, post),
		Moved: []string{}, Skipped: plan.skipped,
		AppliedAt: p.AppliedAt, FollowUpAfter: p.FollowUpAfter, ManualReason: p.ManualReason,
	}
	for _, m := range plan.moves {
		res.Moved = append(res.Moved, rel(l.Root, m.to))
	}
	return res, ClearHandoff(l.Handoff)
}

// checkOptions validates the target, slug, reason and dates.
func checkOptions(o MoveOptions) error {
	if o.To != StatusApplied && o.To != StatusManual {
		return fmt.Errorf("%w: move target must be applied or manual, got %q", ErrIllegalTransition, o.To)
	}
	if o.Slug == "" || o.Slug != filepath.Base(o.Slug) || o.Slug == "." || o.Slug == ".." ||
		strings.ContainsAny(o.Slug, `/\`) {
		return fmt.Errorf("%w: bad slug %q", ErrBadMove, o.Slug)
	}
	if o.To == StatusManual {
		if strings.TrimSpace(o.Reason) == "" || strings.ContainsAny(o.Reason, "\r\n") {
			return fmt.Errorf("%w: manual needs a one-line reason", ErrBadMove)
		}
		return nil
	}
	if o.Reason != "" {
		return fmt.Errorf("%w: a reason is only for manual", ErrBadMove)
	}
	if _, err := time.Parse(time.DateOnly, o.Today); err != nil {
		return fmt.Errorf("%w: today %q is not YYYY-MM-DD", ErrBadMove, o.Today)
	}
	if o.FollowUpDays < 0 {
		return fmt.Errorf("%w: follow_up_days %d is negative", ErrBadMove, o.FollowUpDays)
	}
	return nil
}

// readSource reads and parses the posting and checks the transition. A
// posting already archived is an illegal transition, not a missing file.
func readSource(l Layout, src string, o MoveOptions) (Posting, []byte, os.FileMode, error) {
	data, err := os.ReadFile(src)
	if errors.Is(err, os.ErrNotExist) {
		for _, d := range []string{DirApplied, DirManual} {
			if _, serr := os.Stat(filepath.Join(l.Dir(d), o.Slug+".md")); serr == nil {
				return Posting{}, nil, 0, fmt.Errorf("%w: %s is already in %s/", ErrIllegalTransition, o.Slug, d)
			}
		}
	}
	if err != nil {
		return Posting{}, nil, 0, err
	}
	fi, err := os.Stat(src)
	if err != nil {
		return Posting{}, nil, 0, err
	}
	p, body, err := ParsePosting(src, data)
	if err != nil {
		return Posting{}, nil, 0, err
	}
	if err := Transition(p.Status, o.To); err != nil {
		return Posting{}, nil, 0, fmt.Errorf("%s: %w", src, err)
	}
	return p, body, fi.Mode().Perm(), nil
}

// checkDest refuses when <dest>/<slug>.md or <dest>/<slug>/ exists.
func checkDest(dest, slug string) error {
	for _, p := range []string{filepath.Join(dest, slug+".md"), filepath.Join(dest, slug)} {
		_, err := os.Lstat(p)
		if err == nil {
			return fmt.Errorf("%w: %s", ErrDestExists, p)
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

// setFrontmatter applies ADR 0026's field updates for the target status.
func setFrontmatter(p *Posting, o MoveOptions) error {
	p.Status = o.To
	if o.To == StatusManual {
		p.ManualReason = strings.TrimSpace(o.Reason)
		return nil
	}
	today, err := time.Parse(time.DateOnly, o.Today)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBadMove, err)
	}
	p.AppliedAt = o.Today
	p.FollowUpAfter = today.AddDate(0, 0, o.FollowUpDays).Format(time.DateOnly)
	return nil
}

// rel returns path relative to root, slash-separated; absolute and
// slash-separated when it lies outside root.
func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(r)
}
