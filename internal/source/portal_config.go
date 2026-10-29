package source

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/lamurian/caddie/internal/config"
)

// Portal is one organisation's portal config, portals/<org>.yml (ADR 0019).
//
// Field values in Fields depend on the listing format. For HTML each is a CSS
// selector relative to the item, optionally suffixed `@attr` to read an
// attribute instead of the text (`a.title@href`); an empty selector before
// `@` means the item element itself (`@data-id`). For JSON each is a dotted
// key path inside the item (`location.city`).
type Portal struct {
	Org          string       `yaml:"-"`             // file base name, e.g. "who"
	File         string       `yaml:"-"`             // path it was loaded from, used in errors
	Name         string       `yaml:"name"`          // organisation name, used as Job.Company
	ListingURL   string       `yaml:"listing_url"`   // page or API endpoint listing openings
	Method       string       `yaml:"method"`        // GET (default) or POST
	Body         string       `yaml:"body"`          // request body for POST (JSON)
	ItemSelector string       `yaml:"item_selector"` // CSS selector for one item in HTML
	JSONPath     string       `yaml:"json_path"`     // dotted path to the item array in JSON
	Fields       PortalFields `yaml:"fields"`
	DetailURL    string       `yaml:"detail_url"`  // job page pattern with {id}
	ATS          string       `yaml:"ats"`         // taleo, oracle, successfactors, workday, inspira, bespoke…
	JSRendered   bool         `yaml:"js_rendered"` // listing needs JavaScript: skipped with a notice
	Notes        string       `yaml:"notes"`
}

// PortalFields maps Job fields to selectors (HTML) or key paths (JSON).
type PortalFields struct {
	ID       string `yaml:"id"`
	Title    string `yaml:"title"`
	URL      string `yaml:"url"`
	Location string `yaml:"location"`
	Deadline string `yaml:"deadline"`
	Salary   string `yaml:"salary"`
}

var portalOrg = regexp.MustCompile(`^[a-z0-9-]+$`)

// LoadPortal reads and validates one portal config. Unknown keys are errors.
func LoadPortal(path string) (Portal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Portal{}, fmt.Errorf("%w: %v", config.ErrConfig, err)
	}
	p := Portal{}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&p); err != nil && !errors.Is(err, io.EOF) {
		return Portal{}, fmt.Errorf("%w: %s: %v", config.ErrConfig, path, err)
	}
	p.Org = strings.TrimSuffix(filepath.Base(path), ".yml")
	p.File = path
	p.Method = strings.ToUpper(p.Method)
	if p.Method == "" {
		p.Method = "GET"
	}
	return p, p.Validate()
}

// LoadPortals loads dir/<org>.yml for each org, in order. A missing file is
// a notice, not an error; an invalid org name or config is an error.
func LoadPortals(dir string, orgs []string) ([]Portal, []string, error) {
	var portals []Portal
	var notices []string
	for _, org := range orgs {
		if !portalOrg.MatchString(org) {
			return nil, nil, fmt.Errorf("%w: sources.portals: %q is not a portal name ([a-z0-9-]+)",
				config.ErrConfig, org)
		}
		path := filepath.Join(dir, org+".yml")
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			notices = append(notices, fmt.Sprintf("%s:%s: skipped: %s does not exist", PortalFamily, org, path))
			continue
		}
		p, err := LoadPortal(path)
		if err != nil {
			return nil, nil, err
		}
		portals = append(portals, p)
	}
	return portals, notices, nil
}

// SourceName is the portal's Job.Source and --source value, "portal:<org>".
func (p Portal) SourceName() string { return PortalFamily + ":" + p.Org }

// problem is one validation failure: the offending key and what is wrong.
type problem struct{ field, msg string }

// Validate checks the config; every problem wraps config.ErrConfig and reads
// "<file>: <field>: <problem>". A js_rendered config is skipped at search
// time, so it needs only name, listing_url and ats.
func (p Portal) Validate() error {
	checks := []func() *problem{p.checkBasics, p.checkRequest}
	if !p.JSRendered {
		checks = append(checks, p.checkExtraction, p.checkJSONPaths, p.checkLinks)
	}
	for _, check := range checks {
		if pr := check(); pr != nil {
			return fmt.Errorf("%w: %s: %s: %s", config.ErrConfig, p.File, pr.field, pr.msg)
		}
	}
	return nil
}

func (p Portal) checkBasics() *problem {
	switch {
	case strings.TrimSpace(p.Name) == "":
		return &problem{"name", "required"}
	case p.ListingURL == "":
		return &problem{"listing_url", "required"}
	case !absoluteHTTP(p.ListingURL):
		return &problem{"listing_url", "must be an absolute http(s) URL"}
	case strings.TrimSpace(p.ATS) == "":
		return &problem{"ats", "required (taleo, oracle, successfactors, workday, inspira, bespoke…)"}
	}
	return nil
}

func (p Portal) checkRequest() *problem {
	switch {
	case p.Method != "GET" && p.Method != "POST":
		return &problem{"method", fmt.Sprintf("%q is not one of GET|POST", p.Method)}
	case p.Body != "" && p.Method != "POST":
		return &problem{"body", "only allowed with method: POST"}
	}
	return nil
}

func (p Portal) checkExtraction() *problem {
	switch {
	case p.ItemSelector != "" && p.JSONPath != "":
		return &problem{"item_selector, json_path", "set exactly one, not both"}
	case p.ItemSelector == "" && p.JSONPath == "":
		return &problem{"item_selector, json_path", "set exactly one"}
	case p.Fields.Title == "":
		return &problem{"fields.title", "required"}
	}
	return nil
}

func (p Portal) checkLinks() *problem {
	if p.DetailURL != "" {
		if !absoluteHTTP(p.DetailURL) {
			return &problem{"detail_url", "must be an absolute http(s) URL"}
		}
		if !strings.Contains(p.DetailURL, "{id}") {
			return &problem{"detail_url", "must contain {id}"}
		}
	}
	if p.Fields.URL == "" && (p.Fields.ID == "" || p.DetailURL == "") {
		return &problem{"fields.url", "required unless fields.id and detail_url are set"}
	}
	return nil
}

func absoluteHTTP(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != ""
}

// Hosts lists the sorted, unique hostnames of listing_url and detail_url,
// which seed the browser's egress allowlist (phase 06).
func (p Portal) Hosts() []string {
	var hosts []string
	for _, raw := range []string{p.ListingURL, p.DetailURL} {
		if h := hostOf(raw); h != "" && !slices.Contains(hosts, h) {
			hosts = append(hosts, h)
		}
	}
	slices.Sort(hosts)
	return hosts
}

// Matches reports whether rawURL is one of this portal's job pages: its host
// equals detail_url's and its path and query match the pattern, whose {id}
// is captured and returned. Without detail_url any URL on the listing host
// matches, with an empty id.
func (p Portal) Matches(rawURL string) (string, bool) {
	host := hostOf(rawURL)
	if host == "" {
		return "", false
	}
	if p.DetailURL == "" {
		return "", host == hostOf(p.ListingURL)
	}
	if host != hostOf(p.DetailURL) {
		return "", false
	}
	before, _, _ := strings.Cut(rawURL, "#")
	expr := `^` + strings.ReplaceAll(regexp.QuoteMeta(urlTail(p.DetailURL)), `\{id\}`, `([^/?#&]+)`) + `$`
	re, err := regexp.Compile(expr)
	if err != nil {
		return "", false
	}
	m := re.FindStringSubmatch(urlTail(before))
	if m == nil {
		return "", false
	}
	if len(m) > 1 {
		return m[1], true
	}
	return "", true
}

func hostOf(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}

// urlTail is the part of a URL after its scheme and host: the path and
// query, "/" when both are empty. It works on the raw string so that a
// pattern's {id} is not escaped.
func urlTail(raw string) string {
	_, rest, ok := strings.Cut(raw, "://")
	if !ok {
		return raw
	}
	i := strings.IndexAny(rest, "/?")
	if i < 0 {
		return "/"
	}
	if rest[i] == '?' {
		return "/" + rest[i:]
	}
	return rest[i:]
}
