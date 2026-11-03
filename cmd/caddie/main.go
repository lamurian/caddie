// Command caddie is the deterministic half of Caddie: parsing, checks and
// file mechanics for the agent that drives .agents/skills/caddie/SKILL.md.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/lamurian/caddie/internal/browse"
	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/document"
	"github.com/lamurian/caddie/internal/source"
)

// env is everything a command touches outside its arguments.
type env struct {
	stdin          io.Reader // answers to `caddie init` questions
	stdout, stderr io.Writer
	lookup         config.LookupFunc
	dotenv         string
	lookPath       func(string) (string, error)
	client         source.Doer   // HTTP for search and posting save
	today          func() string // the date as YYYY-MM-DD
	quarto         func(bin string) document.Quarto
	browser        browse.Exec // runs agent-browser for `caddie browse`
}

type command func(args []string, e env) error

var commands = map[string]command{
	"profile":    cmdProfile,
	"search":     cmdSearch,
	"seen":       cmdSeen,
	"posting":    cmdPosting,
	"answers":    cmdAnswers,
	"verify":     cmdVerify,
	"keywords":   cmdKeywords,
	"render":     cmdRender,
	"browse":     cmdBrowse,
	"move":       cmdMove,
	"follow-ups": cmdFollowUps,
	"doctor":     cmdDoctor,
	"init":       cmdInit,
	"validate":   cmdValidate,
}

const usageText = `Usage: caddie <command> [flags]

Commands:
  init [--questions] [--answers <file|->]
                                      first run: ask the setup questions, write
                                      .env, preferences.yml and answers.yml if missing
  validate [preferences|answers]      check both files against their schema
  profile [--json] [--with-contact]   JSON summary of history/ + preferences.yml
  search [--source s] [--limit n] [--include-seen]
                                      JSON jobs from enabled boards and portals
  seen add <id…> --shown|--rejected   record jobs shown to or rejected by the user
  seen list [--rejected]              JSON of seen.jsonl, one entry per job
  posting save <url>                  fetch a posting into postings/<slug>.md
  posting steps <slug> "<step>"…      add steps to a posting's manual checklist
  answers add --key k --scope global|company:<x>|posting:<slug> "<text>"
                                      store an answer in answers.yml
  answers show [--for <slug>] [--form]
                                      JSON answers: global, or merged for a posting
  answers list                        JSON of every stored answer and its scope
  verify <doc.md> [--json]            trace every claim of a draft to history/ and answers
  keywords <posting.md> <doc.md> [--json]
                                      keyword coverage of a draft against a posting
  render resume|cover-letter <slug> [--json]
                                      verified draft to PDF, copied to handoff/
  browse <session> <agent-browser args…>
                                      run agent-browser in the sandbox; session =
                                      portal hostname, uploads only from /handoff
  move <slug> applied|manual [--reason "<one line>"]
                                      archive a posting and its documents (JSON)
  follow-ups [--json]                 follow-ups due, open manual steps, deadlines
  doctor [--json]                     check configuration, workspace and history
  help                                show this help

Exit codes: see cmd/README.md.
`

func main() {
	os.Exit(run(os.Args[1:], env{
		stdin: os.Stdin, stdout: os.Stdout, stderr: os.Stderr, lookup: os.LookupEnv,
		dotenv: ".env", lookPath: exec.LookPath,
		client:  &http.Client{Timeout: 30 * time.Second},
		today:   func() string { return time.Now().Format("2006-01-02") },
		quarto:  func(bin string) document.Quarto { return document.ExecQuarto{Bin: bin} },
		browser: browse.OSExec{},
	}))
}

func run(args []string, e env) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(e.stderr, usageText)
		return exitUsage
	}
	switch args[0] {
	case "help", "-h", "-help", "--help":
		_, _ = fmt.Fprint(e.stdout, usageText)
		return exitOK
	}
	cmd, ok := commands[args[0]]
	if !ok {
		_, _ = fmt.Fprintf(e.stderr, "caddie: unknown command %q\n\n%s", args[0], usageText)
		return exitUsage
	}
	err := cmd(args[1:], e)
	if errors.Is(err, flag.ErrHelp) {
		return exitOK
	}
	if err != nil {
		_, _ = fmt.Fprintf(e.stderr, "caddie %s: %v\n", args[0], err)
	}
	return exitCode(err)
}

// parseFlags parses a subcommand's flags and rejects positional arguments.
func parseFlags(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return err
		}
		return usageErr("%v", err)
	}
	if fs.NArg() > 0 {
		return usageErr("unexpected argument %q", fs.Arg(0))
	}
	return nil
}

// parseArgs parses flags that may appear before, between or after the
// positional arguments, and returns the positional arguments.
func parseArgs(fs *flag.FlagSet, args []string) ([]string, error) {
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil, err
			}
			return nil, usageErr("%v", err)
		}
		if fs.NArg() == 0 {
			return pos, nil
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
}

func newFlags(name string, e env) *flag.FlagSet {
	fs := flag.NewFlagSet("caddie "+name, flag.ContinueOnError)
	fs.SetOutput(e.stderr)
	return fs
}
