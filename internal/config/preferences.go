package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"go.yaml.in/yaml/v3"
)

// ErrNoPreferences means preferences.yml does not exist yet.
var ErrNoPreferences = errors.New("preferences file not found")

// Preferences mirrors preferences.yml (PRD §12.7). Contact is never
// serialised to JSON; callers add it explicitly when asked to.
type Preferences struct {
	Contact   Contact   `yaml:"contact" json:"-"`
	Salary    Salary    `yaml:"salary" json:"salary"`
	WorkMode  []string  `yaml:"work_mode" json:"work_mode"`
	Locations Locations `yaml:"locations" json:"locations"`
	Roles     Roles     `yaml:"roles" json:"roles"`
	Companies Companies `yaml:"companies" json:"companies"`
	Sources   Sources   `yaml:"sources" json:"sources"`
	Ranking   Ranking   `yaml:"ranking" json:"ranking"`
	Apply     Apply     `yaml:"apply" json:"apply"`
}

// Contact holds private details, merged in only when explicitly requested.
type Contact struct {
	Email string `yaml:"email" json:"email"`
	Phone string `yaml:"phone" json:"phone"`
}

// Salary is the salary expectation.
type Salary struct {
	Currency string `yaml:"currency" json:"currency"`
	Period   string `yaml:"period" json:"period"`
	Minimum  int    `yaml:"minimum" json:"minimum"`
	Target   int    `yaml:"target" json:"target"`
}

// Locations groups location preferences.
type Locations struct {
	Preferred  []string `yaml:"preferred" json:"preferred"`
	Acceptable []string `yaml:"acceptable" json:"acceptable"`
	Excluded   []string `yaml:"excluded" json:"excluded"`
}

// Roles lists role keywords to include and exclude.
type Roles struct {
	Include []string `yaml:"include" json:"include"`
	Exclude []string `yaml:"exclude" json:"exclude"`
}

// Companies lists companies never to show.
type Companies struct {
	Exclude []string `yaml:"exclude" json:"exclude"`
}

// Sources configures the job sources.
type Sources struct {
	Adzuna           Adzuna   `yaml:"adzuna" json:"adzuna"`
	Remotive         Toggle   `yaml:"remotive" json:"remotive"`
	Arbeitnow        Toggle   `yaml:"arbeitnow" json:"arbeitnow"`
	GreenhouseBoards []string `yaml:"greenhouse_boards" json:"greenhouse_boards"`
	LeverBoards      []string `yaml:"lever_boards" json:"lever_boards"`
	Portals          []string `yaml:"portals" json:"portals"`
}

// Adzuna configures the Adzuna source.
type Adzuna struct {
	Enabled   bool     `yaml:"enabled" json:"enabled"`
	Countries []string `yaml:"countries" json:"countries"`
}

// Toggle enables or disables a source.
type Toggle struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// Ranking holds the ranking hints given to the agent.
type Ranking struct {
	TopN    int     `yaml:"top_n" json:"top_n"`
	Weights Weights `yaml:"weights" json:"weights"`
}

// Weights are the relative ranking weights.
type Weights struct {
	SkillsMatch float64 `yaml:"skills_match" json:"skills_match"`
	SalaryFit   float64 `yaml:"salary_fit" json:"salary_fit"`
	WorkModeFit float64 `yaml:"work_mode_fit" json:"work_mode_fit"`
	LocationFit float64 `yaml:"location_fit" json:"location_fit"`
}

// Apply holds the application rules. RequireGate must stay true.
type Apply struct {
	RequireGate         bool `yaml:"require_gate" json:"require_gate"`
	FollowUpDays        int  `yaml:"follow_up_days" json:"follow_up_days"`
	DeadlineWarningDays int  `yaml:"deadline_warning_days" json:"deadline_warning_days"`
}

// DefaultPreferences are used for keys absent from preferences.yml.
func DefaultPreferences() Preferences {
	return Preferences{
		Salary:  Salary{Period: "year"},
		Ranking: Ranking{TopN: 10},
		Apply:   Apply{RequireGate: true, FollowUpDays: 21, DeadlineWarningDays: 7},
	}
}

// LoadPreferences reads and validates a preferences file. Unknown keys are
// errors, so typos such as `require_gat` cannot pass silently.
func LoadPreferences(path string) (Preferences, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Preferences{}, fmt.Errorf("%w: %w: %s (copy preferences.example.yml)",
			ErrConfig, ErrNoPreferences, path)
	}
	if err != nil {
		return Preferences{}, fmt.Errorf("%w: %v", ErrConfig, err)
	}
	return ParsePreferences(path, data)
}

// ParsePreferences decodes preferences YAML; name is used in errors.
func ParsePreferences(name string, data []byte) (Preferences, error) {
	p := DefaultPreferences()
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&p); err != nil && !errors.Is(err, io.EOF) {
		return Preferences{}, fmt.Errorf("%w: %s: %v", ErrConfig, name, err)
	}
	if err := p.validate(); err != nil {
		return Preferences{}, fmt.Errorf("%w: %s: %v", ErrConfig, name, err)
	}
	return p, nil
}

var (
	workModes = []string{"wfo", "hybrid", "wfh", "wfa"}
	periods   = []string{"year", "month", "hour"}
)

func (p Preferences) validate() error {
	for _, m := range p.WorkMode {
		if !slices.Contains(workModes, m) {
			return fmt.Errorf("work_mode: %q is not one of wfo|hybrid|wfh|wfa", m)
		}
	}
	if !slices.Contains(periods, p.Salary.Period) {
		return fmt.Errorf("salary.period: %q is not one of year|month|hour", p.Salary.Period)
	}
	if p.Ranking.TopN < 1 {
		return fmt.Errorf("ranking.top_n: must be at least 1")
	}
	return nil
}
