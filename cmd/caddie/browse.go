package main

import (
	"context"
	"strings"

	"github.com/lamurian/caddie/internal/browse"
	"github.com/lamurian/caddie/internal/config"
)

const browseUsage = "want `browse <session> <agent-browser args…>`, session = portal hostname"

// cmdBrowse runs one agent-browser command in the session named after the
// portal's hostname. Everything after the session goes to agent-browser
// verbatim (it has its own flags), so nothing here is parsed as a caddie
// flag. Validation, argv and the timeout live in internal/browse; a
// rejected call never reaches the browser. agent-browser's own output is
// streamed through, so when it exits non-zero its message is already on
// stderr and the error (exit 8) only adds the command that failed.
func cmdBrowse(args []string, e env) error {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return usageErr(browseUsage)
	}
	if len(args) == 1 {
		return usageErr("no agent-browser command after session %q; %s", args[0], browseUsage)
	}
	cfg, err := config.Load(e.dotenv, e.lookup)
	if err != nil {
		return err
	}
	return browse.Browse(context.Background(), cfg, e.browser, args[0], args[1:], e.stdout, e.stderr)
}
