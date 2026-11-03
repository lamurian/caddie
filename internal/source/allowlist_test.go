package source

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"testing"
)

// allowlistFile is the browser proxy's egress allowlist (phase 06, ADR 0002).
const allowlistFile = "../../docker/proxy/allowlist.conf"

// loadAllowlist parses allowlist.conf the way tinyproxy 1.11 does: leading
// whitespace is skipped, a pattern ends at whitespace or an unescaped '#',
// blank and comment lines are dropped. Each pattern is an extended regex
// matched against the host, case-insensitively (FilterType ere).
func loadAllowlist(t *testing.T) []*regexp.Regexp {
	t.Helper()
	f, err := os.Open(allowlistFile)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	var res []*regexp.Regexp
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		pat, _, _ := strings.Cut(fields[0], "#")
		if !strings.HasSuffix(pat, "$") || !strings.Contains(pat, "^") {
			t.Errorf("%s: %q is not anchored at both ends", allowlistFile, pat)
		}
		res = append(res, regexp.MustCompile("(?i)"+pat))
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return res
}

func allowed(list []*regexp.Regexp, host string) bool {
	for _, re := range list {
		if re.MatchString(host) {
			return true
		}
	}
	return false
}

// Every host a portal config contacts is on the allowlist, as is Inspira,
// where UN applications happen (portals/README.md Hosts table).
func TestAllowlistCoversPortalHosts(t *testing.T) {
	list := loadAllowlist(t)
	portals, err := LoadAllPortals("../../portals")
	if err != nil || len(portals) == 0 {
		t.Fatalf("no portal configs loaded: %v", err)
	}
	hosts := []string{"inspira.un.org"}
	for _, p := range portals {
		hosts = append(hosts, p.Hosts()...)
	}
	for _, want := range []string{"careers.who.int", "careers.un.org"} {
		if !strings.Contains(strings.Join(hosts, " "), want) {
			t.Errorf("portal hosts %v lack %s", hosts, want)
		}
	}
	for _, h := range hosts {
		if !allowed(list, h) {
			t.Errorf("%s: host %s is not allowlisted", allowlistFile, h)
		}
	}
}

// The PRD §9 hosts are allowed, wildcards cover subdomains only of their own
// domain, and look-alike or unlisted hosts such as example.com are denied.
func TestAllowlistMatching(t *testing.T) {
	list := loadAllowlist(t)
	allow := []string{
		"boards.greenhouse.io", "job-boards.greenhouse.io", "jobs.lever.co",
		"jobs.ashbyhq.com", "who.wd3.myworkdayjobs.com", "jobs.smartrecruiters.com",
		"careers.who.int", "careers.un.org", "inspira.un.org", "worldbankgroup.csod.com",
		"www.adb.org", "www.isdb.org", "www.mckinsey.com", "careers.bcg.com",
		"www.bain.com", "CAREERS.WHO.INT",
	}
	deny := []string{
		"example.com", "www.example.com", "careers.who.int.example.com",
		"evil-careers.who.int", "careersxwho.int", "myworkdayjobs.com.evil.net",
		"notmyworkdayjobs.com", "un.org", "localhost",
	}
	for _, h := range allow {
		if !allowed(list, h) {
			t.Errorf("%s should be allowed", h)
		}
	}
	for _, h := range deny {
		if allowed(list, h) {
			t.Errorf("%s should be denied", h)
		}
	}
}
