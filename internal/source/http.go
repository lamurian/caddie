package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrFetch, redact(rawURL), err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", accept)
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
		return nil, fmt.Errorf("%w: %s: HTTP %d", ErrFetch, redact(rawURL), resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrFetch, redact(rawURL), err)
	}
	return body, nil
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

// redact drops the query string and any user info from a URL.
func redact(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "(invalid URL)"
	}
	return u.Scheme + "://" + u.Host + u.Path
}
