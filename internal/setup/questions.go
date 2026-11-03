package setup

import "strings"

// Kind says how an answer is read: free text, a whole number, a
// comma-separated list, one option, or an ordered list of options.
type Kind string

// Question kinds.
const (
	KindText    Kind = "text"
	KindInt     Kind = "int"
	KindList    Kind = "list"
	KindChoice  Kind = "choice"
	KindChoices Kind = "choices"
)

// Files a question can belong to.
const (
	FilePreferences = "preferences"
	FileAnswers     = "answers"
)

// Question is one default setup question. ID is the dotted YAML path of
// the answer inside File; Default uses the same text a user would type.
type Question struct {
	ID       string   `json:"id"`
	File     string   `json:"file"`
	Prompt   string   `json:"prompt"`
	Kind     Kind     `json:"kind"`
	Default  string   `json:"default,omitempty"`
	Options  []string `json:"options,omitempty"`
	Required bool     `json:"required,omitempty"`
}

// Questions is the default question set for both files. portals are the
// organisation portal configs available (portals/*.yml); all are offered
// and enabled by default. apply.require_gate is never asked: it is
// always written as true (ADR 0003).
func Questions(portals []string) []Question {
	p := func(id, prompt string, k Kind, def string) Question {
		return Question{ID: id, File: FilePreferences, Prompt: prompt, Kind: k, Default: def}
	}
	a := func(key, prompt, def string) Question {
		return Question{ID: "global." + key, File: FileAnswers, Prompt: prompt, Kind: KindText, Default: def}
	}
	roles := p("roles.include", "Job titles or role keywords to search for (comma-separated)", KindList, "")
	roles.Required = true
	period := p("salary.period", "Salary period", KindChoice, "year")
	period.Options = []string{"year", "month", "hour"}
	modes := p("work_mode", "Work modes you accept, most preferred first", KindChoices, "hybrid, wfh, wfo")
	modes.Options = []string{"wfo", "hybrid", "wfh", "wfa"}
	srcs := p("sources.portals", "Organisation portals to search", KindChoices, strings.Join(portals, ", "))
	srcs.Options = portals
	return []Question{
		p("contact.email", "Email for applications (kept private)", KindText, ""),
		p("contact.phone", "Phone for applications (kept private)", KindText, ""),
		roles,
		p("roles.exclude", "Title words that rule a job out", KindList, ""),
		p("salary.currency", "Salary currency (ISO code)", KindText, "EUR"),
		period,
		p("salary.minimum", "Minimum acceptable salary (0: no minimum)", KindInt, "0"),
		p("salary.target", "Target salary (0: unset)", KindInt, "0"),
		modes,
		p("locations.preferred", "Preferred locations (cities, countries, \"Remote (EU)\")", KindList, ""),
		p("locations.acceptable", "Acceptable locations", KindList, ""),
		p("locations.excluded", "Locations to rule out", KindList, ""),
		p("companies.exclude", "Companies never to show", KindList, ""),
		srcs,
		p("sources.adzuna.countries", "Adzuna country codes (needs keys in .env; empty: off)", KindList, ""),
		p("sources.greenhouse_boards", "Greenhouse board slugs (from boards.greenhouse.io/<slug>)", KindList, ""),
		p("sources.lever_boards", "Lever board slugs (from jobs.lever.co/<slug>)", KindList, ""),
		p("ranking.top_n", "How many ranked jobs to show per search", KindInt, "10"),
		p("apply.follow_up_days", "Days after applying before a follow-up is due", KindInt, "21"),
		p("apply.deadline_warning_days", "Warn this many days before a deadline", KindInt, "7"),
		a("notice_period", "Notice period at your current job (e.g. 1 month)", ""),
		a("right_to_work", "Right to work / visa status for your target countries", ""),
		a("earliest_start", "Earliest start date", ""),
		a("referees", "Referees", "Available on request"),
	}
}

// For keeps the questions of the given files.
func For(qs []Question, files ...string) []Question {
	var out []Question
	for _, q := range qs {
		for _, f := range files {
			if q.File == f {
				out = append(out, q)
			}
		}
	}
	return out
}
