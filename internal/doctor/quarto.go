package doctor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lamurian/caddie/internal/document"
)

// quartoTimeout bounds the test render; a first run warms Quarto's caches.
const quartoTimeout = 2 * time.Minute

// quartoTailLines is how many trailing output lines a failure keeps.
const quartoTailLines = 5

// Files of the test render, inside a fresh temp dir outside the workspace.
const (
	quartoDoc = "doctor.qmd"
	quartoPDF = "doctor.pdf"
	quartoQmd = "---\nformat: typst\n---\n\nCaddie doctor test render.\n"
)

// checkQuarto warns when quarto is missing and fails when it is found but
// cannot render a trivial document, since `caddie render` would fail too.
func checkQuarto(r *Report, opts Options, bin string) {
	path, err := opts.LookPath(bin)
	if err != nil {
		r.add("quarto", StatusWarn, "%s not found on PATH; needed by `caddie render` "+
			"(install Quarto or set CADDIE_QUARTO_BIN)", bin)
		return
	}
	q := document.Quarto(document.ExecQuarto{Bin: path})
	if opts.Quarto != nil {
		q = opts.Quarto(path)
	}
	if err := testRender(q); err != nil {
		r.add("quarto", StatusFail, "found at %s but cannot render a test document "+
			"(fix or reinstall Quarto; `caddie render` needs it): %v", path, err)
		return
	}
	r.add("quarto", StatusOK, "found at %s; rendered a test PDF", path)
}

// testRender renders quartoDoc to quartoPDF in a temp dir it always removes.
func testRender(q document.Quarto) (err error) {
	dir, err := os.MkdirTemp("", "caddie-doctor-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()
	if err := os.WriteFile(filepath.Join(dir, quartoDoc), []byte(quartoQmd), 0o600); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), quartoTimeout)
	defer cancel()
	out, err := q.Run(ctx, dir, "render", quartoDoc, "--to", "typst", "--output", quartoPDF)
	if err != nil {
		return fmt.Errorf("%w%s", err, outputTail(out))
	}
	info, err := os.Stat(filepath.Join(dir, quartoPDF))
	if err != nil || info.Size() == 0 {
		return fmt.Errorf("quarto exited 0 but wrote no %s%s", quartoPDF, outputTail(out))
	}
	return nil
}

// outputTail returns the last non-empty output lines on one line, prefixed
// with ": ", or "" when there was no output.
func outputTail(out []byte) string {
	var lines []string
	for _, l := range strings.Split(string(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return ""
	}
	if len(lines) > quartoTailLines {
		lines = lines[len(lines)-quartoTailLines:]
	}
	return ": " + strings.Join(lines, " | ")
}
