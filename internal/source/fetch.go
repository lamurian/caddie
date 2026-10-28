package source

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Web is the source of a posting saved from a page no board adapter knows.
const Web = "web"

// Detail is one fetched posting: the normalised Job and its description
// HTML, from which the posting's sections are extracted.
type Detail struct {
	Job  Job
	HTML string
}

var (
	greenhousePath = regexp.MustCompile(`^/([^/]+)/jobs/(\d+)`)
	leverPath      = regexp.MustCompile(`^/([^/]+)/([0-9a-fA-F-]{36})`)
)

type ref struct{ source, board, id string }

// parseRef recognises Greenhouse and Lever posting URLs.
func parseRef(u *url.URL) (ref, bool) {
	host := strings.ToLower(u.Hostname())
	if strings.HasSuffix(host, "greenhouse.io") {
		if m := greenhousePath.FindStringSubmatch(u.Path); m != nil {
			return ref{Greenhouse, m[1], m[2]}, true
		}
	}
	if host == "jobs.lever.co" || host == "jobs.eu.lever.co" {
		if m := leverPath.FindStringSubmatch(u.Path); m != nil {
			return ref{Lever, m[1], strings.ToLower(m[2])}, true
		}
	}
	return ref{}, false
}

// ID derives a Job ID from a posting URL: the board's native ID for
// Greenhouse and Lever, else "url:" and a hash of the URL (ADR 0016).
func ID(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil {
		if r, ok := parseRef(u); ok {
			return r.source + ":" + r.id
		}
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(rawURL)))
	return "url:" + hex.EncodeToString(sum[:])[:12]
}

// Fetch retrieves one posting. Greenhouse and Lever URLs go through their
// APIs; any other page is read for schema.org JobPosting data, falling back
// to the page's title and main text. The Job may lack a company or title
// when the page does not state them.
func Fetch(ctx context.Context, c Doer, rawURL string) (Detail, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Detail{}, fmt.Errorf("%w: %q is not an http(s) URL", ErrFetch, rawURL)
	}
	r, ok := parseRef(u)
	switch {
	case ok && r.source == Greenhouse:
		return fetchGreenhouse(ctx, c, r)
	case ok && r.source == Lever:
		return fetchLever(ctx, c, r, u.Hostname())
	}
	return fetchPage(ctx, c, rawURL)
}

func fetchGreenhouse(ctx context.Context, c Doer, r ref) (Detail, error) {
	var x greenhouseJob
	u := fmt.Sprintf("%s/%s/jobs/%s?pay_transparency=true", greenhouseBase, url.PathEscape(r.board), r.id)
	if err := getJSON(ctx, c, u, &x); err != nil {
		return Detail{}, err
	}
	j, _ := x.job(r.board)
	return Detail{Job: j, HTML: x.Content}, nil
}

func fetchLever(ctx context.Context, c Doer, r ref, host string) (Detail, error) {
	base := leverBase
	if host == "jobs.eu.lever.co" {
		base = strings.Replace(base, "api.lever.co", "api.eu.lever.co", 1)
	}
	var x leverJob
	u := fmt.Sprintf("%s/%s/%s?mode=json", base, url.PathEscape(r.board), r.id)
	if err := getJSON(ctx, c, u, &x); err != nil {
		return Detail{}, err
	}
	j, _ := x.job(r.board)
	return Detail{Job: j, HTML: x.HTML()}, nil
}

func fetchPage(ctx context.Context, c Doer, rawURL string) (Detail, error) {
	body, err := get(ctx, c, rawURL, "text/html")
	if err != nil {
		return Detail{}, err
	}
	d := pageDetail(string(body))
	d.Job.ID, d.Job.Source, d.Job.URL = ID(rawURL), Web, rawURL
	if d.Job.Company == "" {
		u, _ := url.Parse(rawURL)
		d.Job.Company = strings.TrimPrefix(u.Hostname(), "www.")
	}
	text := Text(d.HTML)
	if d.Job.WorkMode == "" {
		d.Job.WorkMode = WorkMode(d.Job.Location)
	}
	if d.Job.Deadline == "" {
		d.Job.Deadline = Deadline(text)
	}
	d.Job.Summary = summarise(text)
	return d, nil
}
