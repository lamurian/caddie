package main

import (
	"context"
	"net/url"

	"github.com/lamurian/caddie/internal/memory"
	"github.com/lamurian/caddie/internal/posting"
)

func cmdPosting(args []string, e env) error {
	if len(args) == 0 || args[0] != "save" {
		return usageErr("want `posting save <url>`")
	}
	fs := newFlags("posting save", e)
	fs.Bool("json", true, "print JSON (always on)")
	pos, err := parseArgs(fs, args[1:])
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr("want exactly one URL")
	}
	if u, err := url.Parse(pos[0]); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return usageErr("%q is not an http(s) URL", pos[0])
	}
	_, layout, err := loadLayout(e)
	if err != nil {
		return err
	}
	res, err := posting.Save(context.Background(), pos[0], posting.Options{
		Layout: layout, Store: memory.Store{Dir: layout.State}, Client: e.client, Today: e.today(),
	})
	if err != nil {
		return err
	}
	return writeJSON(e.stdout, res)
}
