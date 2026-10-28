package source

import (
	"github.com/lamurian/caddie/internal/config"
)

// Keys are the credentials a source may need, read from .env.
type Keys struct {
	AdzunaAppID, AdzunaAppKey string
}

// Build returns the sources enabled in preferences.yml, restricted to only
// when it is not empty. A disabled or unconfigured source is not an error:
// it is skipped and explained in notices.
func Build(prefs config.Sources, keys Keys, only string, c Doer) ([]Source, []string) {
	var srcs []Source
	var notices []string
	add := func(s Source, skip string) {
		switch {
		case only != "" && only != s.Name():
		case skip != "":
			notices = append(notices, s.Name()+": skipped: "+skip)
		default:
			srcs = append(srcs, s)
		}
	}
	add(&adzuna{client: c, keys: keys, countries: prefs.Adzuna.Countries}, adzunaSkip(prefs.Adzuna, keys))
	add(&remotive{client: c}, toggleSkip(prefs.Remotive.Enabled))
	add(&arbeitnow{client: c}, toggleSkip(prefs.Arbeitnow.Enabled))
	add(&greenhouse{client: c, boards: prefs.GreenhouseBoards},
		listSkip(prefs.GreenhouseBoards, "sources.greenhouse_boards"))
	add(&lever{client: c, boards: prefs.LeverBoards}, listSkip(prefs.LeverBoards, "sources.lever_boards"))
	return srcs, notices
}

func adzunaSkip(a config.Adzuna, k Keys) string {
	switch {
	case !a.Enabled:
		return "disabled in preferences.yml"
	case k.AdzunaAppID == "" || k.AdzunaAppKey == "":
		return "CADDIE_ADZUNA_APP_ID and CADDIE_ADZUNA_APP_KEY are not set in .env"
	case len(a.Countries) == 0:
		return "sources.adzuna.countries is empty in preferences.yml"
	}
	return ""
}

func toggleSkip(enabled bool) string {
	if enabled {
		return ""
	}
	return "disabled in preferences.yml"
}

func listSkip(list []string, key string) string {
	if len(list) > 0 {
		return ""
	}
	return key + " is empty in preferences.yml"
}
