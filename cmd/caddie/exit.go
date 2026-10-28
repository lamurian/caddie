package main

import (
	"errors"
	"fmt"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/profile"
	"github.com/lamurian/caddie/internal/workspace"
)

// Exit codes: the contract documented in cmd/README.md and mirrored in
// skills/caddie/SKILL.md. Change all three together. Codes 5, 7 and 8 are
// reserved for later phases (not found, refused, external failure).
const (
	exitOK           = 0
	exitCheckFailed  = 1
	exitUsage        = 2
	exitConfig       = 3
	exitMalformed    = 4
	exitIllegalState = 6
	exitInternal     = 9
)

var (
	errUsage       = errors.New("usage error")
	errCheckFailed = errors.New("check failed")
)

func usageErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", errUsage, fmt.Sprintf(format, args...))
}

// exitCodes maps sentinel errors to exit codes; the first match wins.
var exitCodes = []struct {
	err  error
	code int
}{
	{errCheckFailed, exitCheckFailed},
	{errUsage, exitUsage},
	{config.ErrConfig, exitConfig},
	{profile.ErrMalformed, exitMalformed},
	{workspace.ErrBadPosting, exitMalformed},
	{workspace.ErrUnknownStatus, exitMalformed},
	{workspace.ErrIllegalTransition, exitIllegalState},
}

func exitCode(err error) int {
	if err == nil {
		return exitOK
	}
	for _, m := range exitCodes {
		if errors.Is(err, m.err) {
			return m.code
		}
	}
	return exitInternal
}
