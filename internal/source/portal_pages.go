package source

import (
	"context"
	"strconv"
	"strings"
)

// pagePlaceholder marks where the page number goes in listing_url or body
// of a paginated portal (ADR 0028).
const pagePlaceholder = "{page}"

// maxPortalPages caps `pages`, so a config cannot turn one search into a
// crawl of the portal.
const maxPortalPages = 10

// checkPages validates `pages`: 1 to maxPortalPages, with {page} in
// listing_url or body; {page} without `pages` is an error too.
func (p Portal) checkPages() *problem {
	inURL := strings.Contains(p.ListingURL, pagePlaceholder)
	inBody := strings.Contains(p.Body, pagePlaceholder)
	switch {
	case p.Pages < 0 || p.Pages > maxPortalPages:
		return &problem{"pages", "must be between 1 and " + strconv.Itoa(maxPortalPages)}
	case p.Pages > 0 && !inURL && !inBody:
		return &problem{"pages", "needs " + pagePlaceholder + " in listing_url or body"}
	case p.Pages == 0 && inURL:
		return &problem{"listing_url", pagePlaceholder + " is only allowed with pages"}
	case p.Pages == 0 && inBody:
		return &problem{"body", pagePlaceholder + " is only allowed with pages"}
	}
	return nil
}

// listing sends the configured GET or POST and extracts the raw items. A
// paginated portal is asked for pages 1, 2… up to `pages`, stopping at the
// first page that adds no new item (empty, or a repeat of an earlier one,
// as Taleo answers past the end). Any failed page fails the portal.
func (s *portalSource) listing(ctx context.Context) ([]portalItem, error) {
	if s.p.Pages == 0 {
		return s.page(ctx, s.p.ListingURL, s.p.Body)
	}
	var all []portalItem
	seen := map[string]bool{}
	for n := 1; n <= s.p.Pages; n++ {
		num := strconv.Itoa(n)
		items, err := s.page(ctx, strings.ReplaceAll(s.p.ListingURL, pagePlaceholder, num),
			strings.ReplaceAll(s.p.Body, pagePlaceholder, num))
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
