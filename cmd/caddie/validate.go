package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/lamurian/caddie/internal/setup"
	"github.com/lamurian/caddie/internal/workspace"
)

const validateUsage = "want `validate`, `validate preferences` or `validate answers`"

// cmdValidate checks preferences.yml and answers.yml with the loaders every
// command uses (schemas/ documents the same rules for editors). It prints
// one ok/fail line per file; a missing file or bad preferences exits 3,
// malformed answers exit 4.
func cmdValidate(args []string, e env) error {
	pos, err := parseArgs(newFlags("validate", e), args)
	if err != nil {
		return err
	}
	if len(pos) > 1 {
		return usageErr(validateUsage)
	}
	which := ""
	if len(pos) == 1 {
		which = pos[0]
	}
	if which != "" && which != setup.FilePreferences && which != setup.FileAnswers {
		return usageErr("unknown file %q: %s", which, validateUsage)
	}
	_, layout, err := loadLayout(e)
	if err != nil {
		return err
	}
	var errs []error
	for _, c := range checks(layout) {
		if which != "" && which != c.name {
			continue
		}
		errs = append(errs, report(e, c.name, c.path, c.check(c.path)))
	}
	return errors.Join(errs...)
}

type fileCheck struct {
	name, path string
	check      func(string) error
}

func checks(l workspace.Layout) []fileCheck {
	return []fileCheck{
		{setup.FilePreferences, l.Preferences, setup.CheckPreferences},
		{setup.FileAnswers, l.Answers, setup.CheckAnswers},
	}
}

// report prints the ok or fail line for one file and passes err on.
func report(e env, name, path string, err error) error {
	status := "ok  "
	if err != nil {
		status = "fail"
	}
	_, _ = fmt.Fprintf(e.stdout, "%s %s: %s\n", status, name, filepath.Base(path))
	return err
}
