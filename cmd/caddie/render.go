package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/document"
	"github.com/lamurian/caddie/internal/profile"
	"github.com/lamurian/caddie/internal/workspace"
)

// renderTimeout bounds one quarto run.
const renderTimeout = 5 * time.Minute

const renderUsage = "want `render resume|cover-letter <slug>`"

// renderOutput is what `caddie render --json` prints; paths are relative
// to the workspace root where possible.
type renderOutput struct {
	Kind    document.Kind    `json:"kind"`
	Slug    string           `json:"slug"`
	PDF     string           `json:"pdf"`
	Handoff string           `json:"handoff"`
	Status  workspace.Status `json:"status"`
}

// cmdRender renders a verified draft to PDF, copies it to handoff/ and
// advances the posting to `rendered` (ADR 0022). The gate lives in
// document.Render; this is wiring only.
func cmdRender(args []string, e env) error {
	fs := newFlags("render", e)
	asJSON := fs.Bool("json", false, "print the result as JSON")
	pos, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 2 {
		return usageErr(renderUsage)
	}
	kind, err := document.ParseKind(pos[0])
	if err != nil {
		return err
	}
	slug := pos[1]
	if err := checkSlug(slug); err != nil {
		return err
	}
	cfg, layout, prefs, err := loadWorkspace(e)
	if err != nil {
		return err
	}
	in, postingPath, err := renderInput(cfg, layout, kind, slug)
	if err != nil {
		return err
	}
	in.Contact = prefs.Contact
	ctx, cancel := context.WithTimeout(context.Background(), renderTimeout)
	defer cancel()
	res, err := document.Render(ctx, e.quarto(cfg.QuartoBin), in)
	if err != nil {
		return err
	}
	p, err := workspace.AdvanceStatus(postingPath, workspace.StatusRendered)
	if err != nil {
		return err
	}
	out := renderOutput{Kind: kind, Slug: slug, Status: p.Status,
		PDF: wsPath(layout, res.PDF), Handoff: wsPath(layout, res.Handoff)}
	return printRender(e.stdout, out, *asJSON)
}

// renderInput checks everything that can fail before quarto runs: the
// draft, the posting and its status, history/ and the template.
func renderInput(cfg config.Config, layout workspace.Layout, kind document.Kind, slug string) (document.RenderInput, string, error) {
	in := document.RenderInput{Kind: kind, StateDir: layout.State, HandoffDir: layout.Handoff}
	in.Source = filepath.Join(layout.Dir(draftDir(kind)), slug+".md")
	if err := needDraft(layout, in.Source); err != nil {
		return in, "", err
	}
	postingPath, p, err := renderPosting(layout, slug)
	if err != nil {
		return in, "", err
	}
	in.Position, in.Company = p.Title, p.Company
	prof, err := profile.Load(layout.History())
	if err != nil {
		return in, "", err
	}
	in.Identity = prof.Identity
	in.Template, err = renderTemplate(cfg, kind)
	return in, postingPath, err
}

func draftDir(kind document.Kind) string {
	if kind == document.KindCoverLetter {
		return workspace.DirCoverLetters
	}
	return workspace.DirResumes
}

// needDraft fails with exit 5 and the next step when the draft is missing.
func needDraft(layout workspace.Layout, path string) error {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		rel := wsPath(layout, path)
		return fmt.Errorf("%w: no draft %s: write it, then run caddie verify %s", errNotFound, rel, rel)
	}
	return err
}

// renderPosting reads postings/<slug>.md and refuses, before any render,
// a posting that cannot become `rendered`: one in manual/ or applied/
// (exit 6) or one whose status is past it.
func renderPosting(layout workspace.Layout, slug string) (string, workspace.Posting, error) {
	path := filepath.Join(layout.Dir(workspace.DirPostings), slug+".md")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if p, ferr := findPosting(layout, slug); ferr == nil {
			return "", p, fmt.Errorf("%w: posting %q is %s, not in postings/", workspace.ErrIllegalTransition, slug, p.Status)
		}
		return "", workspace.Posting{}, fmt.Errorf("%w: no posting %s", errNotFound, wsPath(layout, path))
	}
	if err != nil {
		return "", workspace.Posting{}, err
	}
	p, _, err := workspace.ParsePosting(path, data)
	if err != nil {
		return "", p, err
	}
	if _, err := workspace.StatusPath(p.Status, workspace.StatusRendered); err != nil {
		return "", p, fmt.Errorf("%s: %w", wsPath(layout, path), err)
	}
	return path, p, nil
}

// renderTemplate resolves the kind's template against the workspace; a
// missing file is a configuration error naming the variable.
func renderTemplate(cfg config.Config, kind document.Kind) (string, error) {
	name, rel := "CADDIE_RESUME_TEMPLATE", cfg.ResumeTemplate
	if kind == document.KindCoverLetter {
		name, rel = "CADDIE_COVER_LETTER_TEMPLATE", cfg.CoverLetterTemplate
	}
	path := cfg.Resolve(rel)
	if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("%w: %s: no template file %s", config.ErrConfig, name, path)
	}
	return path, nil
}

// wsPath is path relative to the workspace root, or path itself when it
// lies outside the workspace.
func wsPath(layout workspace.Layout, path string) string {
	rel, err := document.RecordPath(layout.Root, path)
	if err != nil {
		return path
	}
	return rel
}

func printRender(w io.Writer, out renderOutput, asJSON bool) error {
	if asJSON {
		return writeJSON(w, out)
	}
	_, err := fmt.Fprintf(w, "rendered %s\nhandoff %s\nstatus %s\n", out.PDF, out.Handoff, out.Status)
	return err
}
