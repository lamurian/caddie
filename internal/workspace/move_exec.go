package workspace

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Filesystem operations Move goes through; tests swap them to inject
// failures. rename is used for every artefact move and its undo.
var (
	rename       = os.Rename
	writePosting = writeNew
	removeFile   = os.Remove
)

type stepKind int

const (
	stepDir  stepKind = iota // created directory; undo removes it (if empty)
	stepMove                 // moved file; undo moves it back
	stepFile                 // created file; undo removes it
)

type step struct {
	kind     stepKind
	from, to string
}

func (s step) String() string {
	switch s.kind {
	case stepDir:
		return "create " + s.to
	case stepMove:
		return "move " + s.from + " → " + s.to
	default:
		return "write " + s.to
	}
}

// journal records completed steps so a failure can be undone in reverse.
type journal struct {
	done []step
	dirs map[string]bool
}

// execute runs the plan: create <dest>/<slug>/, move the artefacts, write
// the new posting last, then remove the source posting. Any failure undoes
// what was done.
func execute(plan movePlan, src, post string, content []byte, mode os.FileMode) error {
	j := &journal{dirs: map[string]bool{}}
	if err := j.mkdir(plan.dir); err != nil {
		return j.fail("create "+plan.dir, err)
	}
	for _, m := range plan.moves {
		if err := j.mkdir(filepath.Dir(m.to)); err != nil {
			return j.fail("create "+filepath.Dir(m.to), err)
		}
		if err := moveFile(m.from, m.to); err != nil {
			return j.fail("move "+m.from, err)
		}
		j.done = append(j.done, step{stepMove, m.from, m.to})
	}
	if err := writePosting(post, content, mode); err != nil {
		return j.fail("write "+post, err)
	}
	j.done = append(j.done, step{kind: stepFile, to: post})
	if err := removeFile(src); err != nil {
		return j.fail("remove "+src, err)
	}
	return nil
}

func (j *journal) mkdir(dir string) error {
	if j.dirs[dir] {
		return nil
	}
	if err := os.Mkdir(dir, 0o755); err != nil {
		return err
	}
	j.dirs[dir] = true
	j.done = append(j.done, step{kind: stepDir, to: dir})
	return nil
}

// fail undoes every recorded step in reverse and returns the original
// error, joined with one naming each step that could not be undone.
func (j *journal) fail(what string, cause error) error {
	err := fmt.Errorf("move: %s: %w", what, cause)
	var stuck []string
	for i := len(j.done) - 1; i >= 0; i-- {
		s := j.done[i]
		if uerr := undo(s); uerr != nil {
			stuck = append(stuck, fmt.Sprintf("%s (%v)", s, uerr))
		}
	}
	if len(stuck) == 0 {
		return err
	}
	return errors.Join(err, fmt.Errorf("could not undo: %s", strings.Join(stuck, "; ")))
}

func undo(s step) error {
	switch s.kind {
	case stepMove:
		return moveFile(s.to, s.from)
	default:
		return os.Remove(s.to)
	}
}

// moveFile renames src to dst, copying and removing across devices.
func moveFile(src, dst string) error {
	err := rename(src, dst)
	if err == nil || !errors.Is(err, syscall.EXDEV) {
		return err
	}
	if err := copyFile(src, dst); err != nil {
		return err
	}
	if err := os.Remove(src); err != nil {
		_ = os.Remove(dst)
		return err
	}
	return nil
}

// copyFile copies src to a new file dst with src's permission bits.
func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	fi, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fi.Mode().Perm())
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if err == nil {
		err = out.Sync()
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = os.Remove(dst)
	}
	return err
}

// writeNew writes content to path, which must not exist, via a temporary
// file in the same directory and a rename, with permission bits mode.
func writeNew(path string, content []byte, mode os.FileMode) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("%w: %s", ErrDestExists, path)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_, err = tmp.Write(content)
	if err == nil {
		err = tmp.Chmod(mode)
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
