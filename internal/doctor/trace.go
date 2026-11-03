package doctor

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/lamurian/caddie/internal/workspace"
)

// browserUID is the browser container's `node` user, which writes /trace.
const browserUID = 1000

// checkTrace adds `trace`: <state>/trace is the host source of the
// browser's read-write /trace bind mount (ADR 0024). Compose refuses to
// start without it, so a missing directory fails in compose mode; in local
// mode nothing mounts it and `caddie move` merely has no trace to archive.
func checkTrace(r *Report, l workspace.Layout, mode string) {
	dir := filepath.Join(l.State, "trace")
	show := relTo(l.Root, dir)
	fi, err := os.Stat(dir)
	switch {
	case errors.Is(err, fs.ErrNotExist) && mode == "compose":
		r.add("trace", StatusFail, "%s is missing; `docker compose up` refuses to start "+
			"without it: mkdir -p %s (writable by uid %d, docker/README.md)", show, show, browserUID)
	case errors.Is(err, fs.ErrNotExist):
		r.add("trace", StatusWarn, "%s is missing; CADDIE_BROWSER_MODE=%s does not mount it, "+
			"so `caddie move` archives no trace: mkdir -p %s", show, mode, show)
	case err != nil:
		r.add("trace", StatusFail, "cannot stat %s: %v", show, err)
	case !fi.IsDir():
		r.add("trace", StatusFail, "%s is not a directory; remove it, then mkdir -p %s", show, show)
	case mode == "compose" && !browserMayWrite(fi):
		r.add("trace", StatusWarn, "%s may not be writable by uid %d, so screenshots would fail: "+
			"setfacl -m u:%d:rwx %s (docker/README.md)", show, browserUID, browserUID, show)
	default:
		r.add("trace", StatusOK, "%s present", show)
	}
}

// browserMayWrite is a cheap, conservative guess: true unless, on Linux,
// the directory belongs to another uid and grants no group or other write
// (an ACL entry for uid 1000 shows up as a writable group-class mask).
func browserMayWrite(fi fs.FileInfo) bool {
	uid, ok := ownerUID(fi)
	if !ok || uid == browserUID {
		return true
	}
	return fi.Mode().Perm()&0o022 != 0
}

// relTo shows path relative to root when it lies inside root.
func relTo(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path
	}
	return rel
}
