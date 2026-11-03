//go:build linux

package doctor

import (
	"io/fs"
	"syscall"
)

// ownerUID returns the uid owning fi. Only Linux is checked: Docker Desktop
// on macOS and Windows maps bind-mount ownership itself.
func ownerUID(fi fs.FileInfo) (uint32, bool) {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return st.Uid, true
}
