package doctor

import (
	"fmt"
	"os"
	"strings"

	"github.com/lamurian/caddie/internal/browse"
	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/document"
	"github.com/lamurian/caddie/internal/memory"
	"github.com/lamurian/caddie/internal/profile"
	"github.com/lamurian/caddie/internal/workspace"
)

// Status is the outcome of one check.
type Status string

// Check outcomes. Only StatusFail makes the report unhealthy.
const (
	StatusOK   Status = "ok"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

// Check is one line of the report.
type Check struct {
	Name    string `json:"name"`
	Status  Status `json:"status"`
	Message string `json:"message"`
}

// Report is the full doctor output.
type Report struct {
	OK     bool    `json:"ok"`
	Checks []Check `json:"checks"`
}

// Options carries the inputs, so tests can fake the environment.
type Options struct {
	DotEnv   string
	Lookup   config.LookupFunc
	LookPath func(file string) (string, error)
	// Quarto runs the quarto found at bin; nil runs it for real.
	Quarto func(bin string) document.Quarto
	// Browser runs the docker compose calls of the stack checks; nil runs
	// them for real.
	Browser browse.Exec
}

func (r *Report) add(name string, s Status, format string, args ...any) {
	r.Checks = append(r.Checks, Check{Name: name, Status: s, Message: fmt.Sprintf(format, args...)})
	if s == StatusFail {
		r.OK = false
	}
}

// Run performs every check it can. A config failure stops the run because
// every later check depends on the configuration.
func Run(opts Options) Report {
	r := Report{OK: true}
	cfg, err := config.Load(opts.DotEnv, opts.Lookup)
	if err != nil {
		r.add("config", StatusFail, "%v", err)
		return r
	}
	r.add("config", StatusOK, "settings loaded (workspace %s)", cfg.Workspace)
	layout := workspace.New(cfg)
	checkDirs(&r, layout)
	checkTrace(&r, layout, cfg.BrowserMode)
	checkPreferences(&r, layout)
	checkAnswers(&r, layout)
	checkHistory(&r, layout)
	checkQuarto(&r, opts, cfg.QuartoBin)
	checkBrowser(&r, opts, cfg)
	return r
}

func checkDirs(r *Report, l workspace.Layout) {
	if missing := l.Missing(); len(missing) > 0 {
		r.add("workspace", StatusFail, "missing directories: %s", strings.Join(missing, ", "))
		return
	}
	r.add("workspace", StatusOK, "all workspace directories present")
}

func checkPreferences(r *Report, l workspace.Layout) {
	prefs, err := config.LoadPreferences(l.Preferences)
	if err != nil {
		r.add("preferences", StatusFail, "%v", err)
		r.add("gate", StatusFail, "cannot check apply.require_gate: preferences did not load")
		return
	}
	r.add("preferences", StatusOK, "%s parsed", l.Preferences)
	if !prefs.Apply.RequireGate {
		r.add("gate", StatusFail, "apply.require_gate is false in %s: it must be true; "+
			"the human submit gate cannot be disabled (ADR 0003)", l.Preferences)
		return
	}
	r.add("gate", StatusOK, "apply.require_gate is true")
}

// checkAnswers warns while answers.yml is missing (the agent's first-run
// check or `caddie init` creates it) and fails when it is malformed.
func checkAnswers(r *Report, l workspace.Layout) {
	if _, err := os.Stat(l.Answers); err != nil {
		r.add("answers", StatusWarn, "%s not found: run caddie init", l.Answers)
		return
	}
	if _, err := memory.LoadAnswers(l.Answers); err != nil {
		r.add("answers", StatusFail, "%v", err)
		return
	}
	r.add("answers", StatusOK, "%s parsed", l.Answers)
}

func checkHistory(r *Report, l workspace.Layout) {
	p, err := profile.Load(l.History())
	if err != nil {
		r.add("history", StatusFail, "%v", err)
		return
	}
	r.add("history", StatusOK, "%s parsed: %d sections for %s", l.History(), len(p.Sections), p.Identity.Name)
}
