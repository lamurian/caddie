package document

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/profile"
)

// Kind is the kind of document to render; it is also the word used in the
// PDF name (ADR 0022).
type Kind string

// The document kinds Render knows.
const (
	KindResume      Kind = "resume"
	KindCoverLetter Kind = "cover-letter"
)

var (
	// ErrKind marks a document kind other than resume or cover-letter.
	ErrKind = errors.New("unknown document kind")
	// ErrNotVerified marks a draft whose exact content has no verified
	// hash: it was never verified or was edited since (ADR 0004).
	ErrNotVerified = errors.New("not verified")
	// ErrQuarto marks a failed or unusable quarto run.
	ErrQuarto = errors.New("quarto failed")
)

// ParseKind turns "resume" or "cover-letter" into a Kind.
func ParseKind(s string) (Kind, error) {
	switch k := Kind(s); k {
	case KindResume, KindCoverLetter:
		return k, nil
	}
	return "", fmt.Errorf("%w %q: want resume|cover-letter", ErrKind, s)
}

// RenderInput is everything Render needs. Identity comes from
// history/about.md and Contact from preferences.yml; both reach only the
// temporary render directory (ADR 0008).
type RenderInput struct {
	Kind              Kind
	Source            string // draft path, e.g. <ws>/resumes/<slug>.md; PDF goes next to it
	StateDir          string // .caddie, for IsVerified
	HandoffDir        string
	Template          string // path to the .qmd (config: CADDIE_*_TEMPLATE, resolved)
	Identity          profile.Identity
	Contact           config.Contact
	Position, Company string // the posting's title and company
}

// RenderResult names the rendered PDF next to the source and its copy in
// the handoff directory.
type RenderResult struct{ PDF, Handoff string }

// Render turns a verified draft into a PDF with quarto. It refuses with
// ErrNotVerified, before any temporary directory or quarto call, unless
// the draft's exact content was verified. It renders in a temporary
// directory (removed afterwards), writes the PDF next to Source and copies
// it into HandoffDir, overwriting earlier renders (ADR 0022).
func Render(ctx context.Context, q Quarto, in RenderInput) (RenderResult, error) {
	content, err := os.ReadFile(in.Source)
	if err != nil {
		return RenderResult{}, err
	}
	ok, err := IsVerified(in.StateDir, content)
	if err != nil {
		return RenderResult{}, err
	}
	if !ok {
		return RenderResult{}, fmt.Errorf("%w: %s was never verified or changed since; run caddie verify %s",
			ErrNotVerified, in.Source, in.Source)
	}
	if err := checkTemplate(in.Template); err != nil {
		return RenderResult{}, err
	}
	tmp, err := os.MkdirTemp("", "caddie-render-")
	if err != nil {
		return RenderResult{}, err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	name := PDFName(in.Identity.Name, in.Kind, in.Position, in.Company)
	if err := prepareRenderDir(tmp, in, content); err != nil {
		return RenderResult{}, err
	}
	if err := runQuarto(ctx, q, tmp, filepath.Base(in.Template), name); err != nil {
		return RenderResult{}, err
	}
	return deliver(filepath.Join(tmp, name), in, name)
}

// checkTemplate fails clearly when the .qmd is missing, before quarto
// would fail with a less helpful message.
func checkTemplate(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("template: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("template %s: not a regular file", path)
	}
	return nil
}

// runQuarto renders base (the .qmd) in dir to the PDF name; any failure,
// including a missing output file, wraps ErrQuarto.
func runQuarto(ctx context.Context, q Quarto, dir, base, name string) error {
	out, err := q.Run(ctx, dir, "render", base, "--to", "typst", "--output", name)
	if err != nil {
		return fmt.Errorf("%w: %v%s", ErrQuarto, err, outputTail(out))
	}
	info, err := os.Stat(filepath.Join(dir, name))
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("%w: no %s was produced%s", ErrQuarto, name, outputTail(out))
	}
	return nil
}

// deliver copies the built PDF next to the source and into the handoff
// directory, creating it if needed.
func deliver(built string, in RenderInput, name string) (RenderResult, error) {
	res := RenderResult{
		PDF:     filepath.Join(filepath.Dir(in.Source), name),
		Handoff: filepath.Join(in.HandoffDir, name),
	}
	if err := copyFile(built, res.PDF); err != nil {
		return RenderResult{}, err
	}
	if err := os.MkdirAll(in.HandoffDir, 0o755); err != nil {
		return RenderResult{}, err
	}
	if err := copyFile(built, res.Handoff); err != nil {
		return RenderResult{}, err
	}
	return res, nil
}

// tailLines is how many trailing lines of quarto output an error keeps.
const tailLines = 15

// outputTail returns the last lines of quarto's output, prefixed with a
// newline, or "" when there was none.
func outputTail(out []byte) string {
	s := strings.TrimSpace(string(out))
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) > tailLines {
		lines = lines[len(lines)-tailLines:]
	}
	return "\n" + strings.Join(lines, "\n")
}
