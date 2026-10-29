package source

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lamurian/caddie/internal/config"
)

// LoadAllPortals loads every dir/*.yml in name order, enabled in
// preferences.yml or not: `posting save` recognises any configured portal.
// A missing dir means no portals; a misnamed or invalid file is a config
// error naming it (ADR 0019).
func LoadAllPortals(dir string) ([]Portal, error) {
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	paths, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", config.ErrConfig, dir, err)
	}
	portals := make([]Portal, 0, len(paths))
	for _, path := range paths {
		if org := strings.TrimSuffix(filepath.Base(path), ".yml"); !portalOrg.MatchString(org) {
			return nil, fmt.Errorf("%w: %s: file name must be <org>.yml with <org> matching [a-z0-9-]+",
				config.ErrConfig, path)
		}
		p, err := LoadPortal(path)
		if err != nil {
			return nil, err
		}
		portals = append(portals, p)
	}
	return portals, nil
}

// PortalFor finds the first portal whose job pages include rawURL (see
// Matches) and returns it with the job's ID, "portal:<org>:<id>".
func PortalFor(portals []Portal, rawURL string) (Portal, string, bool) {
	for _, p := range portals {
		if id, ok := p.Matches(rawURL); ok {
			return p, PortalJobID(p, id, rawURL), true
		}
	}
	return Portal{}, "", false
}

// PortalJobID is "portal:<org>:<id>", or, when id is empty, the
// 12-hex-digit URL hash that "url:" IDs use (ADR 0016, 0019).
func PortalJobID(p Portal, id, rawURL string) string {
	if id == "" {
		sum := sha256.Sum256([]byte(strings.TrimSpace(rawURL)))
		id = hex.EncodeToString(sum[:])[:12]
	}
	return p.SourceName() + ":" + id
}

// SourceOf is the source named by a Job ID: "portal:<org>" for
// "portal:<org>:<id>", the board for "<board>:<id>", else "" (a "url:"
// ID, or an unknown prefix).
func SourceOf(id string) string {
	prefix, rest, ok := strings.Cut(id, ":")
	if !ok {
		return ""
	}
	if prefix == PortalFamily {
		org, _, ok := strings.Cut(rest, ":")
		if ok && portalOrg.MatchString(org) {
			return PortalFamily + ":" + org
		}
		return ""
	}
	if Known(prefix) {
		return prefix
	}
	return ""
}
