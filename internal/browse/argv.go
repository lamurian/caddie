package browse

import (
	"fmt"

	"github.com/lamurian/caddie/internal/config"
)

// ContainerBin is the agent-browser binary name inside the compose service.
const ContainerBin = "agent-browser"

// Argv returns the command line for one agent-browser call. It does not
// validate; Browse calls ValidateSession and ValidateArgs first.
//
// The session goes first as agent-browser's global flag `--session <name>`
// (agent-browser.dev/sessions), followed by the caller's args verbatim:
//
//	compose: docker compose -f <ComposeFile> exec -T <ComposeService> agent-browser --session <session> <args…>
//	local:   <BrowserBin> --session <session> <args…>
//
// ComposeFile is resolved against the workspace (config.Resolve). `-T`
// disables TTY allocation so output pipes cleanly to the agent. Any other
// mode wraps config.ErrConfig.
func Argv(cfg config.Config, session string, args []string) ([]string, error) {
	var argv []string
	switch cfg.BrowserMode {
	case "compose":
		argv = []string{"docker", "compose", "-f", cfg.Resolve(cfg.ComposeFile),
			"exec", "-T", cfg.ComposeService, ContainerBin}
	case "local":
		argv = []string{cfg.BrowserBin}
	default:
		return nil, fmt.Errorf("%w: CADDIE_BROWSER_MODE=%q, want compose|local",
			config.ErrConfig, cfg.BrowserMode)
	}
	argv = append(argv, "--session", session)
	return append(argv, args...), nil
}
