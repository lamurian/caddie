package main

import (
	"io"

	"github.com/lamurian/caddie/internal/document"
	"github.com/lamurian/caddie/internal/profile"
)

// cmdKeywords prints the advisory keyword coverage of a draft against a
// posting's requirements. It exits 0 whatever the coverage.
func cmdKeywords(args []string, e env) error {
	fs := newFlags("keywords", e)
	asJSON := fs.Bool("json", false, "print the report as JSON")
	pos, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 2 {
		return usageErr("want `keywords <posting.md> <doc.md>`")
	}
	posting, err := readDoc(pos[0])
	if err != nil {
		return err
	}
	draft, err := readDoc(pos[1])
	if err != nil {
		return err
	}
	_, layout, err := loadLayout(e)
	if err != nil {
		return err
	}
	prof, err := profile.Load(layout.History())
	if err != nil {
		return err
	}
	r := document.Keywords(posting, draft, prof)
	if *asJSON {
		return writeJSON(e.stdout, r)
	}
	_, err = io.WriteString(e.stdout, r.Text())
	return err
}
