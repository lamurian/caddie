package profile

import (
	"errors"
	"fmt"
)

// Profile is the parsed career history. It never holds contact details.
type Profile struct {
	Identity Identity  `json:"identity"`
	About    string    `json:"about"`
	Sections []Section `json:"sections"`
}

// Identity is the frontmatter of history/about.md (PRD §12.2).
type Identity struct {
	Name              string            `yaml:"name" json:"name"`
	Nationality       string            `yaml:"nationality" json:"nationality"`
	WorkAuthorization []string          `yaml:"work_authorization" json:"work_authorization"`
	Location          string            `yaml:"location" json:"location"`
	Links             map[string]string `yaml:"links" json:"links"`
	Languages         []Language        `yaml:"languages" json:"languages"`
}

// Language is a spoken language and its level.
type Language struct {
	Name  string `yaml:"name" json:"name"`
	Level string `yaml:"level" json:"level"`
}

// Section is one history file: Entries for most files, Works for
// publication.md and presentation.md.
type Section struct {
	ID      string  `json:"id"`
	Heading string  `json:"heading"`
	Entries []Entry `json:"entries,omitempty"`
	Works   []Work  `json:"works,omitempty"`
}

// Entry is a dated role, degree or engagement. Start and End are "YYYY-MM"
// or "YYYY"; End is "present" for an ongoing entry.
type Entry struct {
	Title        string   `json:"title"`
	Organisation string   `json:"organisation"`
	Location     string   `json:"location,omitempty"`
	Start        string   `json:"start"`
	End          string   `json:"end"`
	Bullets      []string `json:"bullets,omitempty"`
}

// Work is a publication or presentation.
type Work struct {
	Title   string   `json:"title"`
	Venue   string   `json:"venue"`
	Authors []string `json:"authors"`
	Date    string   `json:"date"`
	Bullets []string `json:"bullets,omitempty"`
}

// ErrMalformed marks history that does not follow the schema.
var ErrMalformed = errors.New("malformed history")

// ParseError locates a schema violation. Line is 1-based; 0 means the
// whole file.
type ParseError struct {
	File string
	Line int
	Msg  string
}

func (e *ParseError) Error() string {
	if e.Line == 0 {
		return fmt.Sprintf("%s: %s", e.File, e.Msg)
	}
	return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Msg)
}

// Unwrap lets callers match any ParseError with errors.Is(err, ErrMalformed).
func (e *ParseError) Unwrap() error { return ErrMalformed }

func errAt(file string, line int, format string, args ...any) error {
	return &ParseError{File: file, Line: line, Msg: fmt.Sprintf(format, args...)}
}
