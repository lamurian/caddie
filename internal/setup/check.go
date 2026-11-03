package setup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/memory"
)

// ErrMissing marks a preferences or answers file that does not exist yet;
// it wraps config.ErrConfig, so the CLI exits 3.
var ErrMissing = fmt.Errorf("%w: file not found", config.ErrConfig)

// Exists reports whether path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CheckPreferences validates preferences.yml as every command loads it,
// and refuses apply.require_gate: false (ADR 0003).
func CheckPreferences(path string) error {
	if !Exists(path) {
		return fmt.Errorf("%w: %s (run caddie init)", ErrMissing, path)
	}
	p, err := config.LoadPreferences(path)
	if err != nil {
		return err
	}
	if !p.Apply.RequireGate {
		return fmt.Errorf("%w: %s: apply.require_gate: must be true; "+
			"the human submit gate cannot be disabled (ADR 0003)", config.ErrConfig, path)
	}
	return nil
}

// CheckAnswers validates answers.yml's structure (ADR 0007).
func CheckAnswers(path string) error {
	if !Exists(path) {
		return fmt.Errorf("%w: %s (run caddie init)", ErrMissing, path)
	}
	_, err := memory.LoadAnswers(path)
	return err
}

// Write creates path with data, mode 0600; it never overwrites a file.
func Write(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("%s already exists: not overwritten", path)
	}
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// CopyIfMissing copies from to to (mode 0600) when to does not exist and
// from does; it reports whether it copied.
func CopyIfMissing(from, to string) (bool, error) {
	if Exists(to) || !Exists(from) {
		return false, nil
	}
	data, err := os.ReadFile(from)
	if err != nil {
		return false, err
	}
	return true, Write(to, data)
}
