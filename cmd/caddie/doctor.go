package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/lamurian/caddie/internal/doctor"
)

func cmdDoctor(args []string, e env) error {
	fs := newFlags("doctor", e)
	asJSON := fs.Bool("json", false, "print the report as JSON")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	r := doctor.Run(doctor.Options{DotEnv: e.dotenv, Lookup: e.lookup, LookPath: e.lookPath})
	if err := printReport(e.stdout, r, *asJSON); err != nil {
		return err
	}
	if !r.OK {
		return fmt.Errorf("%w: %d problem(s) found", errCheckFailed, failures(r))
	}
	return nil
}

func printReport(w io.Writer, r doctor.Report, asJSON bool) error {
	if asJSON {
		return writeJSON(w, r)
	}
	var b strings.Builder
	for _, c := range r.Checks {
		fmt.Fprintf(&b, "%-4s  %-11s  %s\n", c.Status, c.Name, c.Message)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func failures(r doctor.Report) int {
	n := 0
	for _, c := range r.Checks {
		if c.Status == doctor.StatusFail {
			n++
		}
	}
	return n
}
