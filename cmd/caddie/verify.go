package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lamurian/caddie/internal/document"
	"github.com/lamurian/caddie/internal/memory"
	"github.com/lamurian/caddie/internal/profile"
	"github.com/lamurian/caddie/internal/workspace"
)

// verifyOutput is what `caddie verify --json` prints. Hash is set only on
// a pass; the arrays are never null.
type verifyOutput struct {
	Path        string           `json:"path"`
	OK          bool             `json:"ok"`
	Claims      int              `json:"claims"`
	Untraceable []document.Claim `json:"untraceable"`
	Hash        string           `json:"hash,omitempty"`
	Notices     []string         `json:"notices"`
}

func cmdVerify(args []string, e env) error {
	fs := newFlags("verify", e)
	asJSON := fs.Bool("json", false, "print the report as JSON")
	pos, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("want `verify <doc.md>`")
	}
	path := filepath.Clean(pos[0])
	content, err := readDoc(path)
	if err != nil {
		return err
	}
	_, layout, err := loadLayout(e)
	if err != nil {
		return err
	}
	src, notices, err := verifySources(layout, draftSlug(path))
	if err != nil {
		return err
	}
	r := document.Verify(content, src)
	out := verifyOutput{Path: path, OK: r.OK(), Claims: r.Claims, Untraceable: r.Untraceable, Notices: notices}
	if out.OK {
		if out.Hash, err = document.RecordVerified(layout.State, path, content, e.today()); err != nil {
			return err
		}
	}
	if err := printVerify(e.stdout, out, *asJSON); err != nil {
		return err
	}
	if !out.OK {
		return fmt.Errorf("%w: %d untraceable claim(s) in %s", errCheckFailed, len(out.Untraceable), path)
	}
	return nil
}

// readDoc reads a draft or posting named on the command line; a missing
// file is exit 5.
func readDoc(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: no file %q", errNotFound, path)
	}
	return data, err
}

// draftSlug is the posting slug a draft belongs to: its file name without
// ".md" (resumes/<slug>.md, cover-letters/<slug>.md).
func draftSlug(path string) string {
	return strings.TrimSuffix(filepath.Base(path), ".md")
}

// verifySources gathers what the draft may claim: the profile, and the
// answers and targets of its posting. Without a posting only the global
// answers count, and a notice says so.
func verifySources(layout workspace.Layout, slug string) (document.Sources, []string, error) {
	notices := []string{}
	prof, err := profile.Load(layout.History())
	if err != nil {
		return document.Sources{}, nil, err
	}
	a, err := memory.LoadAnswers(layout.Answers)
	if err != nil {
		return document.Sources{}, nil, err
	}
	src := document.Sources{Profile: prof}
	p, err := postingFor(layout, slug)
	switch {
	case errors.Is(err, errNotFound):
		notices = append(notices, fmt.Sprintf("no posting %q: checked against global answers only", slug))
		src.Answers = answerValues(a.For("", ""))
	case err != nil:
		return src, nil, err
	default:
		src.Targets = []string{p.Company, p.Title}
		src.Answers = answerValues(a.For(slug, workspace.Slug(p.Company)))
	}
	return src, notices, nil
}

// postingFor is findPosting for a name that may not be a slug at all.
func postingFor(layout workspace.Layout, slug string) (workspace.Posting, error) {
	if checkSlug(slug) != nil {
		return workspace.Posting{}, errNotFound
	}
	return findPosting(layout, slug)
}

func answerValues(list []memory.Answer) []string {
	out := make([]string, 0, len(list))
	for _, a := range list {
		out = append(out, a.Value)
	}
	return out
}

// printVerify prints the report: each untraceable claim as
// "<path>:<line>: <kind>: <text>", then notes, then one summary line.
func printVerify(w io.Writer, out verifyOutput, asJSON bool) error {
	if asJSON {
		return writeJSON(w, out)
	}
	var b strings.Builder
	for _, c := range out.Untraceable {
		fmt.Fprintf(&b, "%s:%d: %s: %s\n", out.Path, c.Line, c.Kind, c.Text)
	}
	for _, n := range out.Notices {
		fmt.Fprintf(&b, "note: %s\n", n)
	}
	if out.OK {
		fmt.Fprintf(&b, "verified %s: %d claims traced, hash %s\n", out.Path, out.Claims, out.Hash)
	} else {
		fmt.Fprintf(&b, "%s: %d of %d claims untraceable\n", out.Path, len(out.Untraceable), out.Claims)
	}
	_, err := io.WriteString(w, b.String())
	return err
}
