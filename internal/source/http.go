package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Doer sends HTTP requests; *http.Client satisfies it and tests use a fake
// that serves recorded responses from testdata/sources/.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

const (
	userAgent = "caddie (+https://github.com/lamurian/caddie)"
	maxBody   = 8 << 20 // bytes read from any one response
)

// get fetches rawURL and returns the body of a 200 response. Errors wrap
// ErrFetch and never include the query string, which may hold API keys.
func get(ctx context.Context, c Doer, rawURL, accept string) ([]byte, error) {
	return send(ctx, c, http.MethodGet, rawURL, accept, "", nil)
}

// send is get for any method; a non-empty body is sent as JSON. headers (a
// portal's, already validated) are set last, so they may replace Accept or
// Content-Type but never the User-Agent, which validation refuses.
func send(ctx context.Context, c Doer, method, rawURL, accept, body string,
	headers map[string]string) ([]byte, error) {
	var rd io.Reader
	if body != "" {
		rd = strings.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, rd)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrFetch, redact(rawURL), err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", accept)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return nil, fmt.Errorf("%w: %s: %v", ErrFetch, redact(rawURL), err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s: %w", ErrFetch, redact(rawURL), statusError(resp.StatusCode))
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrFetch, redact(rawURL), err)
	}
	return data, nil
}

func getJSON(ctx context.Context, c Doer, rawURL string, v any) error {
	body, err := get(ctx, c, rawURL, "application/json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("%w: %s: unexpected response: %v", ErrFetch, redact(rawURL), err)
	}
	return nil
}

// statusError is an HTTP status other than 200, printed as "HTTP 404".
type statusError int

func (s statusError) Error() string { return fmt.Sprintf("HTTP %d", int(s)) }

// redact drops the query string and any user info from a URL.
func redact(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "(invalid URL)"
	}
	return u.Scheme + "://" + u.Host + u.Path
}
