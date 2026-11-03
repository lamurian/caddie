package main

import (
	"errors"
	"fmt"

	"github.com/lamurian/caddie/internal/browse"
	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/document"
	"github.com/lamurian/caddie/internal/memory"
	"github.com/lamurian/caddie/internal/posting"
	"github.com/lamurian/caddie/internal/profile"
	"github.com/lamurian/caddie/internal/source"
	"github.com/lamurian/caddie/internal/workspace"
)

// Exit codes: the contract documented in cmd/README.md and mirrored in
// skills/caddie/SKILL.md and skills/caddie/references/exit-codes.md.
// Change all four together.
const (
	exitOK           = 0
	exitCheckFailed  = 1
	exitUsage        = 2
	exitConfig       = 3
	exitMalformed    = 4
	exitNotFound     = 5
	exitIllegalState = 6
	exitRefused      = 7
	exitExternal     = 8
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
	{document.ErrKind, exitUsage},
	{config.ErrConfig, exitConfig},
	{profile.ErrMalformed, exitMalformed},
	{workspace.ErrBadPosting, exitMalformed},
	{workspace.ErrUnknownStatus, exitMalformed},
	{memory.ErrMalformedAnswers, exitMalformed},
	{workspace.ErrBadMove, exitUsage},
	{posting.ErrNoSteps, exitUsage},
	{posting.ErrEmptyStep, exitUsage},
	{workspace.ErrIllegalTransition, exitIllegalState},
	{workspace.ErrDestExists, exitIllegalState},
	{posting.ErrNoPosting, exitNotFound},
	{errNotFound, exitNotFound},
	{document.ErrNotVerified, exitRefused},
	{browse.ErrInvalidSession, exitRefused},
	{browse.ErrUploadOutsideHandoff, exitRefused},
	{browse.ErrScreenshotOutsideTrace, exitRefused},
	{browse.ErrBatchRefused, exitRefused},
	{browse.ErrFlagRefused, exitRefused},
	{browse.ErrExec, exitExternal},
	{source.ErrFetch, exitExternal},
	{document.ErrQuarto, exitExternal},
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
