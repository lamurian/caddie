package main

import (
	"github.com/lamurian/caddie/internal/memory"
)

const seenUsage = "want `seen add <id…> --shown|--rejected` or `seen list [--rejected]`"

func cmdSeen(args []string, e env) error {
	if len(args) == 0 {
		return usageErr(seenUsage)
	}
	switch args[0] {
	case "add":
		return seenAdd(args[1:], e)
	case "list":
		return seenList(args[1:], e)
	}
	return usageErr("unknown subcommand %q: %s", args[0], seenUsage)
}

func seenAdd(args []string, e env) error {
	fs := newFlags("seen add", e)
	fs.Bool("json", true, "print JSON (always on)")
	shown := fs.Bool("shown", false, "the jobs were shown to the user")
	rejected := fs.Bool("rejected", false, "the user rejected the jobs")
	ids, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if *shown == *rejected {
		return usageErr("give exactly one of --shown or --rejected")
	}
	if len(ids) == 0 {
		return usageErr("give at least one job ID")
	}
	status := memory.Shown
	if *rejected {
		status = memory.Rejected
	}
	store, err := seenStore(e)
	if err != nil {
		return err
	}
	res, err := store.Mark(ids, status, e.today())
	if err != nil {
		return err
	}
	return writeJSON(e.stdout, res)
}

type seenListOutput struct {
	Entries  []memory.Entry   `json:"entries"`
	Problems []memory.Problem `json:"problems"`
}

func seenList(args []string, e env) error {
	fs := newFlags("seen list", e)
	fs.Bool("json", true, "print JSON (always on)")
	rejected := fs.Bool("rejected", false, "list only rejected jobs")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	store, err := seenStore(e)
	if err != nil {
		return err
	}
	log, err := store.Log()
	if err != nil {
		return err
	}
	out := seenListOutput{Entries: []memory.Entry{}, Problems: log.Problems}
	if out.Problems == nil {
		out.Problems = []memory.Problem{}
	}
	for _, en := range log.Latest() {
		if !*rejected || en.Status == memory.Rejected {
			out.Entries = append(out.Entries, en)
		}
	}
	return writeJSON(e.stdout, out)
}

func seenStore(e env) (memory.Store, error) {
	_, layout, err := loadLayout(e)
	return memory.Store{Dir: layout.State}, err
}
