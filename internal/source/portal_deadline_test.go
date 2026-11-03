package source

import (
	"strings"
	"testing"
)

// deadline_tz converts a full RFC 3339 timestamp to the zone's date (ADR
// 0029); other text, and any text without deadline_tz, keeps findDate.
func TestPortalDeadlineTZ(t *testing.T) {
	ny := Portal{DeadlineTZ: "America/New_York"}
	cases := []struct {
		p          Portal
		text, want string
	}{
		{ny, "2026-11-14T04:59:59.000Z", "2026-11-13"},       // EST: 23:59 on the 13th
		{ny, "2026-10-21T03:59:59.000Z", "2026-10-20"},       // EDT: 23:59 on the 20th
		{ny, " 2026-11-14T12:00:00Z ", "2026-11-14"},         // midday stays on the day
		{ny, "2026-11-14T01:00:00+02:00", "2026-11-13"},      // any offset, not only Z
		{ny, "2026-11-14T04:59:59", "2026-11-14"},            // no offset: date as written
		{ny, "2026-11-14", "2026-11-14"},                     // date only: as written
		{ny, "Closing: 30 November 2026", "2026-11-30"},      // prose: findDate
		{Portal{}, "2026-11-14T04:59:59.000Z", "2026-11-14"}, // no deadline_tz: UTC date
		{ny, "", ""},
	}
	for _, c := range cases {
		if got := c.p.deadline(c.text); got != c.want {
			t.Errorf("%q (tz %q) = %q, want %q", c.text, c.p.DeadlineTZ, got, c.want)
		}
	}
}

// "Local" resolves to the host's zone, so it is refused; UTC and IANA names
// load from the embedded database on any host.
func TestPortalDeadlineTZCheck(t *testing.T) {
	for tz, bad := range map[string]bool{"": false, "UTC": false, "Asia/Manila": false, "Local": true, "EST5": true} {
		pr := Portal{DeadlineTZ: tz}.checkDeadlineTZ()
		if (pr != nil) != bad || pr != nil && (pr.field != "deadline_tz" || !strings.Contains(pr.msg, "zone")) {
			t.Errorf("%q: got %v, want bad=%v", tz, pr, bad)
		}
	}
}
