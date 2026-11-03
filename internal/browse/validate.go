package browse

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

// HandoffDir is where the browser container sees handoff/ (read-only).
const HandoffDir = "/handoff"

const (
	maxSessionLen = 253
	maxLabelLen   = 63
)

// ValidateSession reports whether name is hostname-like: dot-separated
// labels of 1–63 lowercase ASCII letters, digits and hyphens, no label
// starting or ending with a hyphen, at most 253 characters in all.
// Uppercase is rejected rather than folded, so one portal maps to exactly
// one session (and one state file in the browser volume). Errors wrap
// ErrInvalidSession.
func ValidateSession(name string) error {
	if name == "" || len(name) > maxSessionLen {
		return fmt.Errorf("%w: %q: want 1–%d characters", ErrInvalidSession, name, maxSessionLen)
	}
	for _, label := range strings.Split(name, ".") {
		if err := validateLabel(label); err != nil {
			return fmt.Errorf("%w: %q: %w", ErrInvalidSession, name, err)
		}
	}
	return nil
}

func validateLabel(label string) error {
	if label == "" || len(label) > maxLabelLen {
		return errors.New("each dot-separated label needs 1–63 characters")
	}
	if label[0] == '-' || label[len(label)-1] == '-' {
		return errors.New("a label may not start or end with a hyphen")
	}
	for _, r := range label {
		if !isSessionRune(r) {
			return fmt.Errorf("character %q not allowed (want a-z, 0-9, -, .)", r)
		}
	}
	return nil
}

func isSessionRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
}

// ValidateArgs checks the agent-browser arguments that follow the session:
//   - no `--session` flag (either `--session x` or `--session=x`), which
//     would bypass ValidateSession; this wraps ErrInvalidSession;
//   - no bare `batch` token and none of refusedFlags (see refuse.go);
//     these wrap ErrBatchRefused and ErrFlagRefused;
//   - for the first bare `upload` token, `upload <selector> <file…>` needs
//     at least one file and every argument after the selector must pass
//     ValidateUploadPath; this wraps ErrUploadOutsideHandoff;
//   - for the first bare `screenshot` token, every positional after it and
//     any --screenshot-dir value must be under /trace (see trace.go); this
//     wraps ErrScreenshotOutsideTrace.
//
// All checks are deliberately conservative: a bare `batch`, `upload` or
// `screenshot` token anywhere (even as a value) triggers its check.
func ValidateArgs(args []string) error {
	for _, a := range args {
		if a == "--session" || strings.HasPrefix(a, "--session=") {
			return fmt.Errorf("%w: %q: the session is set by caddie, not in the args",
				ErrInvalidSession, a)
		}
	}
	if err := validateRefused(args); err != nil {
		return err
	}
	if i := indexOf(args, "upload"); i >= 0 {
		if err := validateUpload(args[i+1:]); err != nil {
			return err
		}
	}
	if i := indexOf(args, "screenshot"); i >= 0 {
		return validateScreenshot(args[i+1:])
	}
	return nil
}

func indexOf(args []string, token string) int {
	for i, a := range args {
		if a == token {
			return i
		}
	}
	return -1
}

func validateUpload(rest []string) error {
	if len(rest) < 2 {
		return fmt.Errorf("%w: upload needs <selector> <file…> under %s",
			ErrUploadOutsideHandoff, HandoffDir)
	}
	for _, p := range rest[1:] {
		if err := ValidateUploadPath(p); err != nil {
			return err
		}
	}
	return nil
}

// ValidateUploadPath reports whether p, a container path, names a file
// strictly inside /handoff once cleaned: it must be absolute, and
// `/handoff` itself, `/handoffx/…` and `..` escapes such as
// `/handoff/../etc/passwd` are rejected. Errors wrap ErrUploadOutsideHandoff.
func ValidateUploadPath(p string) error {
	if !strings.HasPrefix(p, "/") {
		return fmt.Errorf("%w: %q is not an absolute path", ErrUploadOutsideHandoff, p)
	}
	clean := path.Clean(p)
	if !strings.HasPrefix(clean, HandoffDir+"/") {
		return fmt.Errorf("%w: %q resolves to %q", ErrUploadOutsideHandoff, p, clean)
	}
	return nil
}
