package browse

import (
	"fmt"
	"strings"
)

// batchCommand is agent-browser's command that runs several commands from
// one quoted argument or from stdin. ValidateArgs sees each batched
// command as a single token, so its upload and screenshot checks could
// not run; caddie refuses batch outright.
const batchCommand = "batch"

// refusedFlags are agent-browser 0.39 global flags that would replace what
// caddie and the sandbox fix: the browser binary, profile, saved state,
// CDP target, extensions, Chrome arguments, proxy, provider, config file
// and file:// access (ADR 0002, PRD §10). `--session` has its own check
// in ValidateArgs because it wraps ErrInvalidSession. Short aliases are
// not listed: they clash with command flags, and the sandbox network is
// the backstop.
var refusedFlags = []string{
	"--profile", "--state", "--session-name",
	"--cdp", "--auto-connect", "--executable-path",
	"--extension", "--args", "--proxy", "--proxy-bypass",
	"--provider", "--config", "--allow-file-access",
}

// validateRefused rejects a bare `batch` token anywhere in args
// (ErrBatchRefused) and any refused flag, as `--flag` or `--flag=value`
// (ErrFlagRefused). Like the other checks it is conservative: the same
// token used as a value is refused too.
func validateRefused(args []string) error {
	for _, a := range args {
		if a == batchCommand {
			return fmt.Errorf("%w: run one caddie browse call per command", ErrBatchRefused)
		}
		if f := refusedFlag(a); f != "" {
			return fmt.Errorf("%w: %q: caddie and the sandbox set this, not the args",
				ErrFlagRefused, f)
		}
	}
	return nil
}

// refusedFlag returns the refused flag a names, or "".
func refusedFlag(a string) string {
	name, _, _ := strings.Cut(a, "=")
	for _, f := range refusedFlags {
		if name == f {
			return f
		}
	}
	return ""
}
