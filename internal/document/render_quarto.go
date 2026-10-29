package document

import (
	"context"
	"os/exec"
)

// Quarto runs the quarto binary; tests use a fake.
type Quarto interface {
	// Run runs quarto with args in dir and returns its combined output.
	Run(ctx context.Context, dir string, args ...string) ([]byte, error)
}

// ExecQuarto runs the real quarto binary, Bin ("quarto" when empty).
type ExecQuarto struct{ Bin string }

// Run implements Quarto with exec.CommandContext(Bin, args...) in dir.
func (e ExecQuarto) Run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	bin := e.Bin
	if bin == "" {
		bin = "quarto"
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}
