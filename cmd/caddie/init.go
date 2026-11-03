package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/lamurian/caddie/internal/setup"
	"github.com/lamurian/caddie/internal/workspace"
)

type questionsOutput struct {
	Missing   []string         `json:"missing"`
	Questions []setup.Question `json:"questions"`
}

// cmdInit is the first run: it copies .env.example to .env, then asks the
// default questions for whichever of preferences.yml and answers.yml is
// missing and writes them, validated. Existing files are validated and
// never overwritten. --questions only prints the questions (JSON) for the
// agent; --answers takes the agent's replies as JSON keyed by question id.
func cmdInit(args []string, e env) error {
	fs := newFlags("init", e)
	questions := fs.Bool("questions", false, "print the questions for the missing files as JSON; change nothing")
	answers := fs.String("answers", "", "JSON answers keyed by question id, from a file or - for stdin")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if *questions && *answers != "" {
		return usageErr("--questions and --answers cannot be combined")
	}
	_, layout, err := loadLayout(e)
	if err != nil {
		return err
	}
	missing, err := checkExisting(e, layout, !*questions)
	if err != nil {
		return err
	}
	qs := setup.For(setup.Questions(portalNames(layout)), missing...)
	if *questions {
		return writeJSON(e.stdout, questionsOutput{Missing: missing, Questions: append([]setup.Question{}, qs...)})
	}
	if len(missing) == 0 {
		return nil
	}
	if err := ensureDotEnv(e); err != nil {
		return err
	}
	v, err := initValues(qs, *answers, e)
	if err != nil {
		return err
	}
	return writeMissing(e, layout, missing, v)
}

// checkExisting validates the files that exist (printing a line for each
// when verbose) and returns the names of the missing ones.
func checkExisting(e env, l workspace.Layout, verbose bool) ([]string, error) {
	missing := []string{}
	for _, c := range checks(l) {
		if !setup.Exists(c.path) {
			missing = append(missing, c.name)
			continue
		}
		err := c.check(c.path)
		if verbose || err != nil {
			if err := report(e, c.name, c.path, err); err != nil {
				return nil, err
			}
		}
	}
	return missing, nil
}

func ensureDotEnv(e env) error {
	example := filepath.Join(filepath.Dir(e.dotenv), ".env.example")
	copied, err := setup.CopyIfMissing(example, e.dotenv)
	if copied {
		_, _ = fmt.Fprintf(e.stdout, "created %s from .env.example (add API keys there)\n", e.dotenv)
	}
	return err
}

func initValues(qs []setup.Question, answers string, e env) (setup.Values, error) {
	in := e.stdin
	if in == nil {
		in = strings.NewReader("")
	}
	if answers == "" {
		_, _ = fmt.Fprintln(e.stdout, "Answer each question; Enter takes the [default]; separate list items with commas.")
		return setup.Ask(qs, in, e.stdout)
	}
	if answers != "-" {
		f, err := os.Open(answers)
		if err != nil {
			return nil, fmt.Errorf("%w: --answers: %v", errNotFound, err)
		}
		defer func() { _ = f.Close() }()
		in = f
	}
	var raw map[string]any
	if err := json.NewDecoder(in).Decode(&raw); err != nil && err != io.EOF {
		return nil, usageErr("--answers: want a JSON object keyed by question id: %v", err)
	}
	v, err := setup.FromJSON(qs, raw)
	if err != nil {
		return nil, usageErr("--answers: %v", err)
	}
	return v, nil
}

func writeMissing(e env, l workspace.Layout, missing []string, v setup.Values) error {
	build := map[string]func(setup.Values) ([]byte, error){
		setup.FilePreferences: setup.Preferences, setup.FileAnswers: setup.Answers,
	}
	for _, c := range checks(l) {
		if !slices.Contains(missing, c.name) {
			continue
		}
		data, err := build[c.name](v)
		if err == nil {
			err = setup.Write(c.path, data)
		}
		if err == nil {
			err = c.check(c.path)
		}
		if err := report(e, c.name, c.path, err); err != nil {
			return err
		}
	}
	return nil
}

// portalNames lists the organisation portal configs (portals/*.yml).
func portalNames(l workspace.Layout) []string {
	files, _ := filepath.Glob(filepath.Join(l.Dir(workspace.DirPortals), "*.yml"))
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, strings.TrimSuffix(filepath.Base(f), ".yml"))
	}
	sort.Strings(names)
	return names
}
