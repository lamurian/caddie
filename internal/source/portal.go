package source

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

// portalSource searches one organisation portal (ADR 0019): one request to
// its listing, items extracted by CSS selector or JSON path.
type portalSource struct {
	p      Portal
	client Doer
}

func (s *portalSource) Name() string { return s.p.SourceName() }

// Search fetches the listing and keeps the items whose title matches a
// term. Listings have no description, so only the title is searched, as on
// company boards (ADR 0018).
func (s *portalSource) Search(ctx context.Context, terms []string) ([]Job, error) {
	items, err := s.listing(ctx)
	if err != nil {
		return nil, err
	}
	jobs := []Job{}
	for _, it := range items {
		if !matches(terms, it["title"]) {
			continue
		}
		if j, ok := s.p.job(it); ok {
			jobs = append(jobs, j)
		}
	}
	return jobs, nil
}

// listing sends the configured GET or POST and extracts the raw items.
func (s *portalSource) listing(ctx context.Context) ([]portalItem, error) {
	accept, extract := "application/json", extractJSON
	if s.p.ItemSelector != "" {
		accept, extract = "text/html", extractHTML
	}
	body, err := send(ctx, s.client, s.p.Method, s.p.ListingURL, accept, s.p.Body)
	if err != nil {
		return nil, err
	}
	return extract(s.p, body)
}

// job normalises one item. The URL is the item's, else detail_url with its
// id; an item with neither is dropped. The native ID is the item's id, else
// the {id} detail_url captures from the URL, else a hash of the URL.
func (p Portal) job(it portalItem) (Job, bool) {
	link := it["url"]
	if link == "" && it["id"] != "" && p.DetailURL != "" {
		link = p.detailFor(it["id"])
	}
	if link == "" {
		return Job{}, false
	}
	native := it["id"]
	if native == "" {
		native, _ = p.Matches(link)
	}
	if native == "" {
		sum := sha256.Sum256([]byte(strings.TrimSpace(link)))
		native = hex.EncodeToString(sum[:])[:12]
	}
	j := Job{
		ID: p.SourceName() + ":" + native, Source: p.SourceName(),
		Title: it["title"], Company: p.Name, URL: link, Location: it["location"],
		WorkMode: WorkMode(it["location"], it["title"]), Deadline: findDate(it["deadline"]),
	}
	// A salary field often holds a grade ("P-3"): keep only an amount
	// stated with a currency or a pay period.
	if s := ParseSalary(it["salary"]); s.Currency != "" || s.Period != "" {
		s.apply(&j)
	}
	return j, true
}

// detailFor fills detail_url's {id}, escaped for the query when {id} is in
// it and for the path otherwise.
func (p Portal) detailFor(id string) string {
	esc := url.PathEscape
	if q := strings.Index(p.DetailURL, "?"); q >= 0 && q < strings.Index(p.DetailURL, "{id}") {
		esc = url.QueryEscape
	}
	return strings.ReplaceAll(p.DetailURL, "{id}", esc(id))
}

// knownPortal reports whether name is "portal" or "portal:<org>".
func knownPortal(name string) bool {
	org, ok := strings.CutPrefix(name, PortalFamily+":")
	return name == PortalFamily || ok && portalOrg.MatchString(org)
}

// selects reports whether --source only selects the source called name.
func selects(only, name string) bool {
	return only == "" || only == name || only == PortalFamily && strings.HasPrefix(name, PortalFamily+":")
}

// PortalOrgs lists the orgs in sources.portals that --source only selects,
// in order: the configs search must load.
func PortalOrgs(enabled []string, only string) []string {
	var orgs []string
	for _, org := range enabled {
		if selects(only, PortalFamily+":"+org) {
			orgs = append(orgs, org)
		}
	}
	return orgs
}

// portalSources turns the loaded portals that only selects into sources. A
// js_rendered portal, or a --source portal that preferences.yml does not
// enable, is a notice.
func portalSources(enabled []string, portals []Portal, only string, c Doer) ([]Source, []string) {
	var srcs []Source
	var notices []string
	org, named := strings.CutPrefix(only, PortalFamily+":")
	switch {
	case only == PortalFamily && len(enabled) == 0:
		notices = append(notices, "portal: skipped: sources.portals is empty in preferences.yml")
	case named && !slices.Contains(enabled, org):
		notices = append(notices, only+": skipped: not in sources.portals in preferences.yml")
	}
	for _, p := range portals {
		switch {
		case !selects(only, p.SourceName()):
		case p.JSRendered:
			notices = append(notices, fmt.Sprintf("%s: skipped: listing needs JavaScript (see portals/%s.yml notes)",
				p.SourceName(), p.Org))
		default:
			srcs = append(srcs, &portalSource{p: p, client: c})
		}
	}
	return srcs, notices
}
