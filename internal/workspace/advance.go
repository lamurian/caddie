package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// lifecycle is the forward chain saved → tailored → rendered → applied.
var lifecycle = map[Status]Status{
	StatusSaved:    StatusTailored,
	StatusTailored: StatusRendered,
	StatusRendered: StatusApplied,
}

// StatusPath returns the single legal transitions that take a posting from
// `from` to `to`, walking the lifecycle chain one Transition at a time:
// saved → rendered is [tailored rendered]. from == to returns no steps and
// no error. An unreachable target (applied → rendered, manual → rendered,
// rendered → tailored) returns an error wrapping ErrIllegalTransition; an
// unknown status one wrapping ErrUnknownStatus.
func StatusPath(from, to Status) ([]Status, error) {
	for _, s := range []Status{from, to} {
		if _, err := ParseStatus(string(s)); err != nil {
			return nil, err
		}
	}
	if from == to {
		return nil, nil
	}
	var steps []Status
	for cur := from; ; {
		if Transition(cur, to) == nil {
			return append(steps, to), nil
		}
		next, ok := lifecycle[cur]
		if !ok || Transition(cur, next) != nil {
			return nil, fmt.Errorf("%w: %s → %s", ErrIllegalTransition, from, to)
		}
		steps = append(steps, next)
		cur = next
	}
}

// AdvanceStatus moves the posting file at path (postings/<slug>.md) to
// status `to` along StatusPath and returns the updated posting. The file is
// rewritten atomically as the canonical frontmatter followed by the original
// body byte for byte, keeping its file mode. When the posting already has
// status `to` the file is not touched. Errors wrap ErrBadPosting,
// ErrIllegalTransition or ErrUnknownStatus, or are the os error reading it.
func AdvanceStatus(path string, to Status) (Posting, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Posting{}, err
	}
	p, body, err := ParsePosting(path, data)
	if err != nil {
		return Posting{}, err
	}
	steps, err := StatusPath(p.Status, to)
	if err != nil {
		return Posting{}, fmt.Errorf("%s: %w", path, err)
	}
	if len(steps) == 0 {
		return p, nil
	}
	p.Status = to
	content := append(p.Frontmatter(), body...)
	if err := writeAtomic(path, content); err != nil {
		return Posting{}, err
	}
	return p, nil
}

// writeAtomic replaces path with content via a temporary file in the same
// directory and a rename, keeping path's permission bits.
func writeAtomic(path string, content []byte) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_, err = tmp.Write(content)
	if err == nil {
		err = tmp.Chmod(fi.Mode().Perm())
	}
	if err == nil {
		err = tmp.Sync()
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	if err != nil {
		_ = os.Remove(name)
	}
	return err
}
