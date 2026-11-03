package browse

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/lamurian/caddie/internal/config"
)

// Sentinel errors; the CLI maps them to exit codes.
var (
	// ErrInvalidSession: the session name is not hostname-like, or the
	// arguments try to set their own --session (exit 7).
	ErrInvalidSession = errors.New("invalid browse session")
	// ErrUploadOutsideHandoff: an upload names a file outside /handoff
	// (exit 7).
	ErrUploadOutsideHandoff = errors.New("upload path outside /handoff")
	// ErrExec: agent-browser could not run or exited non-zero (exit 8).
	ErrExec = errors.New("agent-browser failed")
	// ErrTimeout: CADDIE_BROWSE_TIMEOUT expired. It always comes wrapped
	// together with ErrExec, so errors.Is(err, ErrExec) holds too.
	ErrTimeout = errors.New("agent-browser timed out")
)

// Exec runs a browser command line; tests use a fake.
type Exec interface {
	// Run runs argv (argv[0] is the program) until it exits or ctx ends,
	// streaming its output to stdout and stderr.
	Run(ctx context.Context, argv []string, stdout, stderr io.Writer) error
}

// waitDelay bounds how long Run waits for output pipes after a kill.
const waitDelay = 5 * time.Second

// OSExec runs argv with os/exec.
type OSExec struct{}

// Run implements Exec with exec.CommandContext(argv[0], argv[1:]...).
func (OSExec) Run(ctx context.Context, argv []string, stdout, stderr io.Writer) error {
	if len(argv) == 0 {
		return errors.New("empty argv")
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...) // #nosec G204 -- argv validated by Browse
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.WaitDelay = waitDelay
	return cmd.Run()
}

// Browse validates session and args, builds the argv for cfg.BrowserMode,
// and runs it through ex under a cfg.BrowseTimeout deadline. Validation
// failures return before ex is called.
func Browse(ctx context.Context, cfg config.Config, ex Exec, session string,
	args []string, stdout, stderr io.Writer) error {
	if err := ValidateSession(session); err != nil {
		return err
	}
	if err := ValidateArgs(args); err != nil {
		return err
	}
	argv, err := Argv(cfg, session, args)
	if err != nil {
		return err
	}
	if cfg.BrowseTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.BrowseTimeout)
		defer cancel()
	}
	err = ex.Run(ctx, argv, stdout, stderr)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("%w: %w after %s", ErrExec, ErrTimeout, cfg.BrowseTimeout)
	}
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrExec, argv[0], err)
	}
	return nil
}
