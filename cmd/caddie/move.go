package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lamurian/caddie/internal/document"
	"github.com/lamurian/caddie/internal/profile"
	"github.com/lamurian/caddie/internal/workspace"
)

const moveUsage = "want `move <slug> applied|manual [--reason \"<one line>\"]`"

// cmdMove archives a posting and its artefacts into applied/ or manual/
// (ADR 0026). The checks, moves and rollback live in workspace.Move; this
// is wiring only. It never touches the browser.
func cmdMove(args []string, e env) error {
	fs := newFlags("move", e)
	fs.Bool("json", true, "print JSON (always on)")
	reason := fs.String("reason", "", "why the posting goes to manual/ (one line; manual only)")
	pos, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 2 {
		return usageErr(moveUsage)
	}
	slug, to := pos[0], workspace.Status(pos[1])
	if err := checkSlug(slug); err != nil {
		return err
	}
	if to != workspace.StatusApplied && to != workspace.StatusManual {
		return usageErr("move target must be applied or manual, got %q", pos[1])
	}
	_, layout, prefs, err := loadWorkspace(e)
	if err != nil {
		return err
	}
	opts := workspace.MoveOptions{
		Slug: slug, To: to, Reason: *reason, Today: e.today(),
		FollowUpDays: prefs.Apply.FollowUpDays,
		TraceDir:     filepath.Join(layout.State, "trace"),
	}
	if err := movePDFNames(layout, &opts); err != nil {
		return err
	}
	res, err := workspace.Move(layout, opts)
	if res.Slug == "" {
		return err
	}
	if werr := writeJSON(e.stdout, res); werr != nil {
		return werr
	}
	if err != nil {
		// The move succeeded; only clearing handoff/ failed. Report it as
		// internal (exit 9), never as a sentinel the agent might act on.
		return fmt.Errorf("moved, but %s", err.Error())
	}
	return nil
}

// movePDFNames names the PDFs to archive from history/'s identity and the
// posting's title and company. A posting not in postings/ but in applied/
// or manual/ is left for workspace.Move to refuse (exit 6); one nowhere is
// exit 5.
func movePDFNames(layout workspace.Layout, o *workspace.MoveOptions) error {
	path := filepath.Join(layout.Dir(workspace.DirPostings), o.Slug+".md")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if _, ferr := findPosting(layout, o.Slug); ferr == nil {
			return nil
		}
		return fmt.Errorf("%w: no posting %s", errNotFound, wsPath(layout, path))
	}
	if err != nil {
		return err
	}
	p, _, err := workspace.ParsePosting(path, data)
	if err != nil {
		return err
	}
	prof, err := profile.Load(layout.History())
	if err != nil {
		return err
	}
	name := prof.Identity.Name
	o.ResumePDF = document.PDFName(name, document.KindResume, p.Title, p.Company)
	o.LetterPDF = document.PDFName(name, document.KindCoverLetter, p.Title, p.Company)
	return nil
}
