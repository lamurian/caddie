package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

// ErrConfig marks any configuration problem (exit code 3 in the CLI).
var ErrConfig = errors.New("config error")

// LookupFunc reads an environment variable, like os.LookupEnv.
type LookupFunc func(key string) (string, bool)

// Config holds the CADDIE_* settings. Relative paths other than Workspace
// are resolved against Workspace by Resolve.
type Config struct {
	Workspace           string
	Preferences         string
	Answers             string
	StateDir            string
	HandoffDir          string
	BrowserMode         string
	BrowserBin          string
	ComposeFile         string
	ComposeService      string
	BrowseTimeout       time.Duration
	QuartoBin           string
	ResumeTemplate      string
	CoverLetterTemplate string
	AdzunaAppID         string
	AdzunaAppKey        string
	LogLevel            string
	LogFormat           string
}

// Defaults are the built-in values; they match .env.example (tested).
var Defaults = map[string]string{
	"CADDIE_WORKSPACE":             ".",
	"CADDIE_PREFERENCES":           "preferences.yml",
	"CADDIE_ANSWERS":               "answers.yml",
	"CADDIE_STATE_DIR":             ".caddie",
	"CADDIE_HANDOFF_DIR":           "handoff",
	"CADDIE_BROWSER_MODE":          "compose",
	"CADDIE_BROWSER_BIN":           "agent-browser",
	"CADDIE_COMPOSE_FILE":          "docker/compose.yml",
	"CADDIE_COMPOSE_SERVICE":       "browser",
	"CADDIE_BROWSE_TIMEOUT":        "60s",
	"CADDIE_QUARTO_BIN":            "quarto",
	"CADDIE_RESUME_TEMPLATE":       "templates/resume.qmd",
	"CADDIE_COVER_LETTER_TEMPLATE": "templates/cover-letter.qmd",
	"CADDIE_ADZUNA_APP_ID":         "",
	"CADDIE_ADZUNA_APP_KEY":        "",
	"CADDIE_LOG_LEVEL":             "info",
	"CADDIE_LOG_FORMAT":            "text",
}

var allowed = map[string][]string{
	"CADDIE_BROWSER_MODE": {"compose", "local"},
	"CADDIE_LOG_LEVEL":    {"debug", "info", "warn", "error"},
	"CADDIE_LOG_FORMAT":   {"text", "json"},
}

// Load resolves every setting with the precedence env > .env > default.
// dotenvPath may name a file that does not exist. A variable that ends up
// empty falls back to its default.
func Load(dotenvPath string, lookup LookupFunc) (Config, error) {
	file, err := ReadDotEnv(dotenvPath)
	if err != nil {
		return Config{}, err
	}
	if err := checkKeys(dotenvPath, file); err != nil {
		return Config{}, err
	}
	vals := make(map[string]string, len(Defaults))
	for key, def := range Defaults {
		v, ok := lookup(key)
		if !ok {
			v = file[key]
		}
		if v == "" {
			v = def
		}
		vals[key] = v
	}
	return fromValues(vals)
}

func checkKeys(path string, file map[string]string) error {
	var unknown []string
	for key := range file {
		if _, ok := Defaults[key]; !ok && strings.HasPrefix(key, "CADDIE_") {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)
	return fmt.Errorf("%w: %s: unknown setting(s) %s (see .env.example)",
		ErrConfig, path, strings.Join(unknown, ", "))
}

func fromValues(v map[string]string) (Config, error) {
	for key, opts := range allowed {
		if !slices.Contains(opts, v[key]) {
			return Config{}, fmt.Errorf("%w: %s=%q, want one of %s",
				ErrConfig, key, v[key], strings.Join(opts, "|"))
		}
	}
	timeout, err := time.ParseDuration(v["CADDIE_BROWSE_TIMEOUT"])
	if err != nil || timeout <= 0 {
		return Config{}, fmt.Errorf("%w: CADDIE_BROWSE_TIMEOUT=%q is not a positive duration",
			ErrConfig, v["CADDIE_BROWSE_TIMEOUT"])
	}
	return Config{
		Workspace:           v["CADDIE_WORKSPACE"],
		Preferences:         v["CADDIE_PREFERENCES"],
		Answers:             v["CADDIE_ANSWERS"],
		StateDir:            v["CADDIE_STATE_DIR"],
		HandoffDir:          v["CADDIE_HANDOFF_DIR"],
		BrowserMode:         v["CADDIE_BROWSER_MODE"],
		BrowserBin:          v["CADDIE_BROWSER_BIN"],
		ComposeFile:         v["CADDIE_COMPOSE_FILE"],
		ComposeService:      v["CADDIE_COMPOSE_SERVICE"],
		BrowseTimeout:       timeout,
		QuartoBin:           v["CADDIE_QUARTO_BIN"],
		ResumeTemplate:      v["CADDIE_RESUME_TEMPLATE"],
		CoverLetterTemplate: v["CADDIE_COVER_LETTER_TEMPLATE"],
		AdzunaAppID:         v["CADDIE_ADZUNA_APP_ID"],
		AdzunaAppKey:        v["CADDIE_ADZUNA_APP_KEY"],
		LogLevel:            v["CADDIE_LOG_LEVEL"],
		LogFormat:           v["CADDIE_LOG_FORMAT"],
	}, nil
}

// Resolve returns p unchanged if absolute, else joined to the workspace.
func (c Config) Resolve(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(c.Workspace, p)
}
