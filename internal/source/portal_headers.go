package source

import (
	"maps"
	"slices"
	"strings"

	"golang.org/x/net/http/httpguts"
)

// forbiddenHeaders are request headers a portal config may never set: they
// identify the client, carry credentials or belong to the transport. The
// headers key is for functional request parameters only (ADR 0019).
var forbiddenHeaders = []string{
	"user-agent", "authorization", "proxy-authorization", "cookie", "set-cookie",
	"host", "content-length", "forwarded",
}

// secretWords mark a header name that probably carries a credential.
var secretWords = []string{"token", "auth", "key", "secret"}

// checkHeaders rejects headers that spoof the client, carry credentials or
// could inject other headers. Names are checked in sorted order, so the
// error is the same on every run.
func (p Portal) checkHeaders() *problem {
	for _, name := range slices.Sorted(maps.Keys(p.Headers)) {
		if msg := headerProblem(name, p.Headers[name]); msg != "" {
			return &problem{"headers." + name, msg}
		}
	}
	return nil
}

func headerProblem(name, value string) string {
	lower := strings.ToLower(name)
	switch {
	case !httpguts.ValidHeaderFieldName(name):
		return "is not a valid header name"
	case slices.Contains(forbiddenHeaders, lower) || strings.HasPrefix(lower, "x-forwarded-"):
		return "may not be set by a portal config"
	case slices.ContainsFunc(secretWords, func(w string) bool { return strings.Contains(lower, w) }):
		return "looks like a credential; portal configs are public and never hold secrets"
	case strings.ContainsAny(value, "\r\n"):
		return "value contains CR or LF"
	case !httpguts.ValidHeaderFieldValue(value):
		return "value is not a valid header value"
	}
	return ""
}
