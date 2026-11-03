package source

import (
	"strings"
	"time"
	// Embedded zone database: deadline_tz resolves the same on every host,
	// including Windows and minimal containers without /usr/share/zoneinfo.
	_ "time/tzdata"
)

// checkDeadlineTZ validates the optional deadline_tz (ADR 0029): an IANA zone
// name such as America/New_York. "Local" is refused because it depends on
// the host running caddie.
func (p Portal) checkDeadlineTZ() *problem {
	if p.DeadlineTZ == "" {
		return nil
	}
	if p.DeadlineTZ == "Local" {
		return &problem{"deadline_tz", `"Local" depends on the host; name an IANA zone (America/New_York)`}
	}
	if _, err := time.LoadLocation(p.DeadlineTZ); err != nil {
		return &problem{"deadline_tz", p.DeadlineTZ + " is not an IANA time zone (America/New_York)"}
	}
	return nil
}

// deadline reads the deadline field as YYYY-MM-DD. When deadline_tz is set
// and the whole field is an RFC 3339 timestamp with an offset
// ("2026-11-14T04:59:59.000Z"), the instant is converted to that zone, so a
// UN closing time of 23:59 New York gives the posted date rather than the
// next UTC day. Any other text keeps its first unambiguous date (findDate).
func (p Portal) deadline(text string) string {
	if p.DeadlineTZ != "" {
		loc, err := time.LoadLocation(p.DeadlineTZ)
		t, perr := time.Parse(time.RFC3339, strings.TrimSpace(text))
		if err == nil && perr == nil {
			return t.In(loc).Format("2006-01-02")
		}
	}
	return findDate(text)
}
