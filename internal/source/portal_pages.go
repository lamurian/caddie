package source

import (
	"context"
	"strconv"
	"strings"
)

// pagePlaceholder marks where the page number goes in listing_url or body
// of a paginated portal (ADR 0028); offsetPlaceholder marks an item offset
// instead, for APIs that page by `from` (ADR 0030).
const (
	pagePlaceholder   = "{page}"
	offsetPlaceholder = "{offset}"
)

// maxPortalPages caps `pages`, so a config cannot turn one search into a
// crawl of the portal; maxPageSize bounds `page_size`.
const (
	maxPortalPages = 10
	maxPageSize    = 1000
)

// placeholderIn names the request part (listing_url, then body) holding
// the placeholder ph, or "" if neither does.
func (p Portal) placeholderIn(ph string) string {
	switch {
	case strings.Contains(p.ListingURL, ph):
		return "listing_url"
	case strings.Contains(p.Body, ph):
		return "body"
	}
	return ""
}

// checkPages validates `pages` (1 to maxPortalPages) and `page_size` (1 to
// maxPageSize): `pages` needs exactly one of {page} or {offset} in
// listing_url or body, {offset} needs `page_size`, and neither placeholder
// nor `page_size` is allowed alone.
func (p Portal) checkPages() *problem {
	page, offset := p.placeholderIn(pagePlaceholder), p.placeholderIn(offsetPlaceholder)
	switch {
	case p.Pages < 0 || p.Pages > maxPortalPages:
		return &problem{"pages", "must be between 1 and " + strconv.Itoa(maxPortalPages)}
	case p.PageSize < 0 || p.PageSize > maxPageSize:
		return &problem{"page_size", "must be between 1 and " + strconv.Itoa(maxPageSize)}
	}
	if pr := p.checkPlaceholders(page, offset); pr != nil {
		return pr
	}
	switch {
	case p.PageSize > 0 && offset == "":
		return &problem{"page_size", "is only allowed with " + offsetPlaceholder}
	case offset != "" && p.PageSize == 0:
		return &problem{"page_size", "required with " + offsetPlaceholder}
	}
	return nil
}

// checkPlaceholders checks that `pages` comes with exactly one placeholder,
// given the request part holding each ("" if none).
func (p Portal) checkPlaceholders(page, offset string) *problem {
	switch {
	case page != "" && offset != "":
		return &problem{offset, "use " + pagePlaceholder + " or " + offsetPlaceholder + ", not both"}
	case p.Pages == 0 && page != "":
		return &problem{page, pagePlaceholder + " is only allowed with pages"}
	case p.Pages == 0 && offset != "":
		return &problem{offset, offsetPlaceholder + " is only allowed with pages"}
	case p.Pages > 0 && page == "" && offset == "":
		return &problem{"pages", "needs " + pagePlaceholder + " or " + offsetPlaceholder + " in listing_url or body"}
	}
	return nil
}

// fill puts request n (counting from 0) into s: {page} becomes n+1 and
// {offset} n*page_size. A config holds only one of the two.
func (p Portal) fill(s string, n int) string {
	s = strings.ReplaceAll(s, pagePlaceholder, strconv.Itoa(n+1))
	return strings.ReplaceAll(s, offsetPlaceholder, strconv.Itoa(n*p.PageSize))
}

// listing sends the configured GET or POST and extracts the raw items. A
// paginated portal is asked for `pages` requests at most ({page} 1, 2… or
// {offset} 0, page_size…), stopping at the first that adds no new item
// (empty, or a repeat of an earlier one, as Taleo answers past the end).
// Any failed page fails the portal.
func (s *portalSource) listing(ctx context.Context) ([]portalItem, error) {
	if s.p.Pages == 0 {
		return s.page(ctx, s.p.ListingURL, s.p.Body)
	}
	var all []portalItem
	seen := map[string]bool{}
	for n := 0; n < s.p.Pages; n++ {
		items, err := s.page(ctx, s.p.fill(s.p.ListingURL, n), s.p.fill(s.p.Body, n))
		if err != nil {
			return nil, err
		}
		fresh := 0
		for _, it := range items {
			key := it["id"] + "\x00" + it["url"] + "\x00" + it["title"]
			if !seen[key] {
				seen[key] = true
				all = append(all, it)
				fresh++
			}
		}
		if fresh == 0 {
			break
		}
	}
	return all, nil
}

// page sends one request and extracts its items.
func (s *portalSource) page(ctx context.Context, link, body string) ([]portalItem, error) {
	accept, extract := "application/json", extractJSON
	if s.p.ItemSelector != "" {
		accept, extract = "text/html", extractHTML
	}
	data, err := send(ctx, s.client, s.p.Method, link, accept, body, s.p.Headers)
	if err != nil {
		return nil, err
	}
	return extract(s.p, data)
}
