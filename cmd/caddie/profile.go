package main

import (
	"encoding/json"
	"io"

	"github.com/lamurian/caddie/internal/config"
	"github.com/lamurian/caddie/internal/profile"
	"github.com/lamurian/caddie/internal/workspace"
)

// profileOutput is what `caddie profile` prints. Contact is set only with
// --with-contact; Preferences never serialises its own contact block.
type profileOutput struct {
	profile.Profile
	Contact     *config.Contact    `json:"contact,omitempty"`
	Preferences config.Preferences `json:"preferences"`
}

func cmdProfile(args []string, e env) error {
	fs := newFlags("profile", e)
	fs.Bool("json", true, "print JSON (always on for profile)")
	withContact := fs.Bool("with-contact", false, "include email and phone from preferences.yml")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	cfg, err := config.Load(e.dotenv, e.lookup)
	if err != nil {
		return err
	}
	layout := workspace.New(cfg)
	hist, err := profile.Load(layout.History())
	if err != nil {
		return err
	}
	prefs, err := config.LoadPreferences(layout.Preferences)
	if err != nil {
		return err
	}
	out := profileOutput{Profile: hist, Preferences: prefs}
	if *withContact {
		out.Contact = &prefs.Contact
	}
	return writeJSON(e.stdout, out)
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
