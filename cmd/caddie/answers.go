package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lamurian/caddie/internal/memory"
	"github.com/lamurian/caddie/internal/workspace"
)

const answersUsage = "want `answers add --key k --scope s \"<text>\"`, " +
	"`answers show [--for <slug>] [--form]` or `answers list`"

// errNotFound marks an unknown slug or file (exit 5).
var errNotFound = errors.New("not found")

var answerKey = regexp.MustCompile(`^[a-z0-9_]+$`)

func cmdAnswers(args []string, e env) error {
	if len(args) == 0 {
		return usageErr(answersUsage)
	}
	switch args[0] {
	case "add":
		return answersAdd(args[1:], e)
	case "show":
		return answersShow(args[1:], e)
	case "list":
		return answersList(args[1:], e)
	}
	return usageErr("unknown subcommand %q: %s", args[0], answersUsage)
}

type addOutput struct {
	Scope    string `json:"scope"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	Replaced bool   `json:"replaced"`
}

func answersAdd(args []string, e env) error {
	fs := newFlags("answers add", e)
	fs.Bool("json", true, "print JSON (always on)")
	key := fs.String("key", "", "answer key: lowercase letters, digits and _")
	rawScope := fs.String("scope", "", "global, company:<name> or posting:<slug>")
	pos, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if !answerKey.MatchString(*key) {
		return usageErr("--key %q: want lowercase letters, digits and _", *key)
	}
	scope, err := answerScope(*rawScope)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("give exactly one answer text, quoted")
	}
	text := strings.TrimSpace(pos[0])
	if text == "" {
		return usageErr("the answer text is empty")
	}
	_, layout, err := loadLayout(e)
	if err != nil {
		return err
	}
	a, err := memory.LoadAnswers(layout.Answers)
	if err != nil {
		return err
	}
	replaced := a.Set(scope, *key, text)
	if err := a.Save(layout.Answers); err != nil {
		return err
	}
	return writeJSON(e.stdout, addOutput{Scope: scope.String(), Key: *key, Value: text, Replaced: replaced})
}

// answerScope parses --scope. A company name is stored as its slug, the
// key `answers show --for` looks up; a posting must be named by its slug.
func answerScope(raw string) (memory.Scope, error) {
	s, err := memory.ParseScope(raw)
	if err != nil {
		return s, usageErr("--scope: %v", err)
	}
	switch s.Kind {
	case memory.ScopeCompany:
		if s.Name = workspace.Slug(s.Name); s.Name == "" {
			return s, usageErr("--scope %q: the company name has no letters or digits", raw)
		}
	case memory.ScopePosting:
		if err := checkSlug(s.Name); err != nil {
			return s, err
		}
	}
	return s, nil
}

func checkSlug(slug string) error {
	if slug == "" || workspace.Slug(slug) != slug {
		return usageErr("%q is not a posting slug", slug)
	}
	return nil
}

type showOutput struct {
	For     string          `json:"for,omitempty"`
	Company string          `json:"company,omitempty"`
	Answers []memory.Answer `json:"answers"`
}

func answersShow(args []string, e env) error {
	fs := newFlags("answers show", e)
	fs.Bool("json", true, "print JSON (always on)")
	slug := fs.String("for", "", "merge the answers for this posting slug")
	form := fs.Bool("form", false, "leave out cover-letter-only keys (why_org, why_role, lead_story)")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if *slug != "" {
		if err := checkSlug(*slug); err != nil {
			return err
		}
	}
	_, layout, err := loadLayout(e)
	if err != nil {
		return err
	}
	a, err := memory.LoadAnswers(layout.Answers)
	if err != nil {
		return err
	}
	out := showOutput{For: *slug}
	if *slug != "" {
		p, err := findPosting(layout, *slug)
		if err != nil {
			return err
		}
		out.Company = workspace.Slug(p.Company)
	}
	out.Answers = a.For(out.For, out.Company)
	if *form {
		out.Answers = memory.FormAnswers(out.Answers)
	}
	return writeJSON(e.stdout, out)
}

// findPosting reads <slug>.md from postings/, else manual/, else applied/.
func findPosting(layout workspace.Layout, slug string) (workspace.Posting, error) {
	for _, dir := range []string{workspace.DirPostings, workspace.DirManual, workspace.DirApplied} {
		path := filepath.Join(layout.Dir(dir), slug+".md")
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return workspace.Posting{}, err
		}
		p, _, err := workspace.ParsePosting(path, data)
		return p, err
	}
	return workspace.Posting{}, fmt.Errorf("%w: no posting %q in postings/, manual/ or applied/", errNotFound, slug)
}

type listOutput struct {
	Answers []memory.Answer `json:"answers"`
}

func answersList(args []string, e env) error {
	fs := newFlags("answers list", e)
	fs.Bool("json", true, "print JSON (always on)")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	_, layout, err := loadLayout(e)
	if err != nil {
		return err
	}
	a, err := memory.LoadAnswers(layout.Answers)
	if err != nil {
		return err
	}
	return writeJSON(e.stdout, listOutput{Answers: a.All()})
}
