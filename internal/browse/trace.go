package browse

import (
	"fmt"
	"path"
	"strings"
)

// TraceDir is where the browser container sees .caddie/trace/ (read-write),
// the only place a screenshot may be written (ADR 0024).
const TraceDir = "/trace"

// dirFlag is agent-browser's screenshot output-directory flag; its value
// must be TraceDir or inside it.
const dirFlag = "--screenshot-dir"

// valueFlags are screenshot flags whose next argument is a value, not a
// path (agent-browser 0.39.2 README). Any other flag is assumed to take no
// value, so an unknown flag's value is checked as a path and refused
// unless it is under /trace: conservative by design.
var valueFlags = map[string]bool{
	"--threshold":          true,
	"--screenshot-format":  true,
	"--screenshot-quality": true,
}

// validateScreenshot checks the arguments after a bare `screenshot` token:
// every positional and any --screenshot-dir value must be under /trace.
// No positional at all is allowed (agent-browser's temporary default).
func validateScreenshot(rest []string) error {
	for i := 0; i < len(rest); i++ {
		a := rest[i]
		switch {
		case a == dirFlag:
			if i+1 >= len(rest) {
				return fmt.Errorf("%w: %s needs a directory under %s",
					ErrScreenshotOutsideTrace, dirFlag, TraceDir)
			}
			i++
			if err := validateTraceDir(rest[i]); err != nil {
				return err
			}
		case strings.HasPrefix(a, dirFlag+"="):
			if err := validateTraceDir(strings.TrimPrefix(a, dirFlag+"=")); err != nil {
				return err
			}
		case valueFlags[a]:
			i++
		case strings.HasPrefix(a, "-"):
			// A flag without a value, such as --full or --annotate.
		default:
			if err := ValidateScreenshotPath(a); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateTraceDir accepts /trace itself or a directory inside it.
func validateTraceDir(p string) error {
	if strings.HasPrefix(p, "/") && path.Clean(p) == TraceDir {
		return nil
	}
	return ValidateScreenshotPath(p)
}

// ValidateScreenshotPath reports whether p, a container path, names a file
// strictly inside /trace once cleaned: it must be absolute, and `/trace`
// itself, `/tracex/…`, `/handoff/…` and `..` escapes such as
// `/trace/../etc/x.png` are rejected. Errors wrap ErrScreenshotOutsideTrace.
func ValidateScreenshotPath(p string) error {
	if !strings.HasPrefix(p, "/") {
		return fmt.Errorf("%w: %q is not an absolute path", ErrScreenshotOutsideTrace, p)
	}
	clean := path.Clean(p)
	if !strings.HasPrefix(clean, TraceDir+"/") {
		return fmt.Errorf("%w: %q resolves to %q", ErrScreenshotOutsideTrace, p, clean)
	}
	return nil
}
