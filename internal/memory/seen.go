package memory

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Status is what happened to a job the agent showed or the user picked.
type Status string

// Seen statuses (ADR 0007).
const (
	Shown    Status = "shown"
	Rejected Status = "rejected"
	Saved    Status = "saved"
)

func (s Status) valid() bool { return s == Shown || s == Rejected || s == Saved }

// Entry is one line of seen.jsonl (PRD §12.5).
type Entry struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Title   string `json:"title"`
	Company string `json:"company"`
	Status  Status `json:"status"`
	At      string `json:"at"`
}

// Problem is a line of seen.jsonl that could not be used as written.
type Problem struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Message string `json:"message"`
}

func (p Problem) String() string { return fmt.Sprintf("%s:%d: %s", p.File, p.Line, p.Message) }

// Log is seen.jsonl as read: every usable entry in file order, and a
// Problem for every line that was blank, malformed or doubtful.
type Log struct {
	Entries  []Entry
	Problems []Problem
}

// ReadLog reads a seen.jsonl file leniently. A missing file is an empty
// log. Blank and malformed lines are skipped and reported, never dropped
// silently; an entry with an unknown status still counts as seen.
func ReadLog(path string) (Log, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Log{}, nil
	}
	if err != nil {
		return Log{}, err
	}
	var l Log
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for n := 1; sc.Scan(); n++ {
		e, msg, ok := parseLine(sc.Text())
		if msg != "" {
			l.Problems = append(l.Problems, Problem{File: path, Line: n, Message: msg})
		}
		if ok {
			l.Entries = append(l.Entries, e)
		}
	}
	return l, sc.Err()
}

func parseLine(line string) (e Entry, problem string, ok bool) {
	if strings.TrimSpace(line) == "" {
		return e, "blank line, skipped", false
	}
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		return e, fmt.Sprintf("not a JSON object, skipped (%v)", err), false
	}
	e.ID = strings.TrimSpace(e.ID)
	if e.ID == "" {
		return e, "no id, skipped", false
	}
	if !e.Status.valid() {
		return e, fmt.Sprintf("%s: unknown status %q (want shown|rejected|saved), still counted as seen",
			e.ID, e.Status), true
	}
	return e, "", true
}

// Latest returns one entry per ID, holding that ID's last status, in the
// order IDs first appear.
func (l Log) Latest() []Entry {
	index := map[string]int{}
	var out []Entry
	for _, e := range l.Entries {
		if i, ok := index[e.ID]; ok {
			out[i] = e
			continue
		}
		index[e.ID] = len(out)
		out = append(out, e)
	}
	return out
}

// IDs is the set of every seen ID, whatever its status.
func (l Log) IDs() map[string]bool {
	ids := make(map[string]bool, len(l.Entries))
	for _, e := range l.Entries {
		ids[e.ID] = true
	}
	return ids
}

// ByURL returns the last entry recorded for url.
func (l Log) ByURL(url string) (Entry, bool) {
	for i := len(l.Entries) - 1; i >= 0; i-- {
		if l.Entries[i].URL == url {
			return l.Entries[i], true
		}
	}
	return Entry{}, false
}

// appendEntries adds entries as JSON lines, creating the directory and
// file if needed. A hand-edited file missing its final newline gets one
// first, so the new line never fuses with the last.
func appendEntries(path string, entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if needsNewline(f) {
		buf.WriteByte('\n')
	}
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			_ = f.Close()
			return err
		}
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func needsNewline(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil || fi.Size() == 0 {
		return false
	}
	last := make([]byte, 1)
	if _, err := f.ReadAt(last, fi.Size()-1); err != nil {
		return false
	}
	return last[0] != '\n'
}
