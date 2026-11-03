//go:build !linux

package doctor

import "io/fs"

// ownerUID is not checked outside Linux (see trace_linux.go).
func ownerUID(fs.FileInfo) (uint32, bool) { return 0, false }
