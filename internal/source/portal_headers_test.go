package source

import (
	"errors"
	"strings"
	"testing"

	"github.com/lamurian/caddie/internal/config"
)

func TestPortalHeadersRefused(t *testing.T) {
	cases := map[string][2]string{ // name: value, the problem reported
		"User-Agent":          {"Mozilla/5.0", "may not be set"},
		"Authorization":       {"Bearer x", "may not be set"},
		"proxy-authorization": {"Basic x", "may not be set"},
		"Cookie":              {"a=b", "may not be set"},
		"Set-Cookie":          {"a=b", "may not be set"},
		"Host":                {"example.org", "may not be set"},
		"Content-Length":      {"3", "may not be set"},
		"Forwarded":           {"for=1.2.3.4", "may not be set"},
		"X-Forwarded-For":     {"1.2.3.4", "may not be set"},
		"X-Api-Key":           {"abc", "looks like a credential"},
		"X-CSRF-Token":        {"abc", "looks like a credential"},
		"X-Client-SECRET":     {"abc", "looks like a credential"},
		"X-Auth-User":         {"abc", "looks like a credential"},
		"tz":                  {"GMT\r\nX-Evil: 1", "value contains CR or LF"},
		"Bad Name":            {"x", "is not a valid header name"},
	}
	for name, c := range cases {
		p := Portal{File: "portals/x.yml", Name: "X", ListingURL: "https://x.example.org/",
			Method: "GET", ATS: "taleo", JSRendered: true, Headers: map[string]string{name: c[0]}}
		err := p.Validate()
		want := "portals/x.yml: headers." + name + ": " + c[1]
		if !errors.Is(err, config.ErrConfig) || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v, want %q", name, err, want)
		}
	}
}

// Allowed headers are sent after caddie's own, so they may replace Accept;
// the User-Agent stays caddie's.
func TestPortalHeadersSent(t *testing.T) {
	p := Portal{Org: "hdr", Name: "Hdr", ListingURL: "https://x.example.org/api", Method: "GET",
		JSONPath: "items", Fields: PortalFields{Title: "t", URL: "u"}, ATS: "bespoke",
		Headers: map[string]string{"tz": "GMT+00:00", "Accept": "application/vnd.x+json"}}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	d := &portalDoer{t: t, routes: map[string]string{p.ListingURL: `<`}}
	_, _ = (&portalSource{p: p, client: d}).listing(t.Context())
	h := d.reqs[0].Header
	if h.Get("tz") != "GMT+00:00" || h.Get("Accept") != "application/vnd.x+json" || h.Get("User-Agent") != userAgent {
		t.Errorf("headers = %v", h)
	}
}
