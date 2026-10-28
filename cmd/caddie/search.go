package main

import (
	"context"
	"strings"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/memory"
	"github.com/lamurian/caddie/internal/source"
	"github.com/lamurian/caddie/internal/workspace"
)

// defaultLimit is how many jobs `caddie search` prints unless told otherwise.
const defaultLimit = 50

func cmdSearch(args []string, e env) error {
	fs := newFlags("search", e)
	fs.Bool("json", true, "print JSON (always on for search)")
	only := fs.String("source", "", "query only this source: "+strings.Join(source.Names, "|"))
	limit := fs.Int("limit", defaultLimit, "print at most n jobs (0 = no limit)")
	includeSeen := fs.Bool("include-seen", false, "also show jobs already in .caddie/seen.jsonl")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if *only != "" && !source.Known(*only) {
		return usageErr("unknown source %q (want %s)", *only, strings.Join(source.Names, "|"))
	}
	if *limit < 0 {
		return usageErr("--limit must be 0 or more")
	}
	cfg, layout, prefs, err := loadWorkspace(e)
	if err != nil {
		return err
	}
	store := memory.Store{Dir: layout.State}
	log, err := store.Log()
	if err != nil {
		return err
	}
	keys := source.Keys{AdzunaAppID: cfg.AdzunaAppID, AdzunaAppKey: cfg.AdzunaAppKey}
	srcs, notices := source.Build(prefs.Sources, keys, *only, e.client)
	f := source.Filter{ExcludeRoles: prefs.Roles.Exclude, ExcludeCompanies: prefs.Companies.Exclude, Limit: *limit}
	if !*includeSeen {
		f.Seen = log.IDs()
	}
	res, err := source.Search(context.Background(), srcs, prefs.Roles.Include, f)
	if err != nil {
		return err
	}
	for _, p := range log.Problems {
		notices = append(notices, p.String())
	}
	res.Notices = append(notices, res.Notices...)
	if err := store.Remember(res.Jobs); err != nil {
		return err
	}
	return writeJSON(e.stdout, res)
}

// loadWorkspace loads the settings, the workspace layout and preferences.
func loadWorkspace(e env) (config.Config, workspace.Layout, config.Preferences, error) {
	cfg, layout, err := loadLayout(e)
	if err != nil {
		return cfg, layout, config.Preferences{}, err
	}
	prefs, err := config.LoadPreferences(layout.Preferences)
	return cfg, layout, prefs, err
}

// loadLayout loads the settings and the workspace layout only.
func loadLayout(e env) (config.Config, workspace.Layout, error) {
	cfg, err := config.Load(e.dotenv, e.lookup)
	if err != nil {
		return cfg, workspace.Layout{}, err
	}
	return cfg, workspace.New(cfg), nil
}
