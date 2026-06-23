// Package api is a generic, resilient client for the n8n public REST API
// (https://docs.n8n.io/api/). It implements auth, retries with idempotency-aware
// backoff, client-side rate limiting, dry-run curl rendering, and a generic
// Resource[T] so adding a resource is a struct plus an accessor — no shared-code edits.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jjuanrivvera/n8n-cli/internal/version"
)

// apiKeyHeader is the HTTP header name n8n uses for public API authentication.
// (This is a header name, not a secret value.)
const apiKeyHeader = "X-N8N-API-KEY" //nolint:gosec // G101: header name, not a credential

// errDryRun is returned (and swallowed by callers) when a request is short-circuited
// by --dry-run after the equivalent curl has been printed.
var errDryRun = errors.New("dry-run: request not sent")

// IsDryRun reports whether err is the sentinel returned in --dry-run mode.
func IsDryRun(err error) bool { return errors.Is(err, errDryRun) }

// HTTPDoer is the subset of *http.Client we depend on, to ease testing.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a configured n8n API client. Construct it with New and Options.
type Client struct {
	baseURL   string
	apiKey    string
	userAgent string

	httpClient   HTTPDoer
	rateLimiter  *RateLimiter
	retryPolicy  *RetryPolicy
	logger       *slog.Logger
	defaultLimit int

	dryRun       bool
	showToken    bool
	dryRunWriter io.Writer
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL sets the API base URL, e.g. https://n8n.example.com/api/v1.
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = normalizeBaseURL(u) } }

// WithAPIKey sets the n8n API key sent in the X-N8N-API-KEY header.
func WithAPIKey(k string) Option { return func(c *Client) { c.apiKey = k } }

// WithHTTPClient overrides the underlying HTTP doer (used by tests).
func WithHTTPClient(h HTTPDoer) Option { return func(c *Client) { c.httpClient = h } }

// WithLogger sets the structured logger.
func WithLogger(l *slog.Logger) Option { return func(c *Client) { c.logger = l } }

// WithRequestsPerSecond sets the client-side rate limit (<=0 disables limiting).
func WithRequestsPerSecond(rps float64) Option {
	return func(c *Client) { c.rateLimiter = NewRateLimiter(rps, c.logger) }
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) Option { return func(c *Client) { c.userAgent = ua } }

// WithDefaultLimit sets the default page size for list calls.
func WithDefaultLimit(n int) Option { return func(c *Client) { c.defaultLimit = n } }

// WithDryRun enables dry-run mode: requests print an equivalent curl to w and are
// not sent. The API key is redacted unless showToken is true.
func WithDryRun(enabled, showToken bool, w io.Writer) Option {
	return func(c *Client) {
		c.dryRun = enabled
		c.showToken = showToken
		c.dryRunWriter = w
	}
}

// New builds a Client from options, filling in production defaults.
func New(opts ...Option) *Client {
	c := &Client{
		userAgent:    version.UserAgent(),
		logger:       slog.Default(),
		defaultLimit: DefaultLimit,
		dryRunWriter: io.Discard,
	}
	for _, o := range opts {
		o(c)
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	if c.rateLimiter == nil {
		c.rateLimiter = NewRateLimiter(5, c.logger)
	}
	if c.retryPolicy == nil {
		c.retryPolicy = DefaultRetryPolicy(c.logger)
	}
	return c
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string { return c.baseURL }

// DefaultLimit returns the configured default page size.
func (c *Client) DefaultLimit() int { return c.defaultLimit }

// normalizeBaseURL trims a trailing slash and appends /api/v1 when the caller
// supplied only the instance host. This lets users configure either form.
func normalizeBaseURL(u string) string {
	u = strings.TrimRight(strings.TrimSpace(u), "/")
	if u == "" {
		return u
	}
	if !strings.Contains(u, "/api/v") {
		u += "/api/v1"
	}
	return u
}

// buildURL joins the base URL, path, and query into an absolute URL string.
func (c *Client) buildURL(path string, query url.Values) string {
	p := path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	full := c.baseURL + p
	if len(query) > 0 {
		full += "?" + query.Encode()
	}
	return full
}

// newRequest constructs an authenticated *http.Request.
func (c *Client) newRequest(ctx context.Context, method, rawURL string, body []byte) (*http.Request, error) {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.apiKey != "" {
		req.Header.Set(apiKeyHeader, c.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// do executes a request with rate limiting and idempotency-aware retries, and
// returns the response body bytes. In dry-run mode it prints curl and returns errDryRun.
func (c *Client) do(ctx context.Context, method, rawURL string, body []byte) ([]byte, error) {
	if c.dryRun {
		c.printCurl(method, rawURL, body)
		return nil, errDryRun
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryPolicy.MaxRetries; attempt++ {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, err
		}

		req, err := c.newRequest(ctx, method, rawURL, body)
		if err != nil {
			return nil, err
		}

		c.logger.Debug("http request", "method", method, "url", rawURL, "attempt", attempt+1)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if c.retryPolicy.shouldRetry(method, nil, err) && attempt < c.retryPolicy.MaxRetries {
				if serr := c.sleep(ctx, c.retryPolicy.backoff(attempt, nil)); serr != nil {
					return nil, serr
				}
				continue
			}
			return nil, fmt.Errorf("request to %s failed: %w", rawURL, err)
		}

		respBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("reading response from %s: %w", rawURL, readErr)
		}

		c.logger.Debug("http response", "method", method, "url", rawURL, "status", resp.StatusCode, "bytes", len(respBody))
		if resp.StatusCode == http.StatusTooManyRequests {
			c.rateLimiter.Throttle()
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.rateLimiter.Restore()
			return respBody, nil
		}

		apiErr := newAPIError(resp.StatusCode, respBody)
		lastErr = apiErr
		if c.retryPolicy.shouldRetry(method, resp, nil) && attempt < c.retryPolicy.MaxRetries {
			c.logger.Debug("retrying request", "method", method, "url", rawURL,
				"status", resp.StatusCode, "attempt", attempt+1)
			if serr := c.sleep(ctx, c.retryPolicy.backoff(attempt, resp)); serr != nil {
				return nil, serr
			}
			continue
		}
		return nil, apiErr
	}
	return nil, lastErr
}

// sleep waits for d or until ctx is cancelled, returning ctx.Err() on
// cancellation so a retry loop short-circuits instead of running another attempt.
func (c *Client) sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Do performs a request and unmarshals a JSON response into out (when non-nil).
// body may be a []byte/json.RawMessage, a struct/map (marshalled), or nil.
// This is the low-level escape hatch used by the `api` command and Resource[T].
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	raw, err := encodeBody(body)
	if err != nil {
		return err
	}
	respBody, err := c.do(ctx, method, c.buildURL(path, query), raw)
	if err != nil {
		return err
	}
	if out == nil || len(bytes.TrimSpace(respBody)) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

// doRaw performs a request and returns the raw response bytes (for list decoding
// and the `api` command, which prints the body verbatim).
func (c *Client) doRaw(ctx context.Context, method, path string, query url.Values, body any) ([]byte, error) {
	raw, err := encodeBody(body)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, method, c.buildURL(path, query), raw)
}

// encodeBody converts a body argument into JSON bytes (nil stays nil).
func encodeBody(body any) ([]byte, error) {
	switch b := body.(type) {
	case nil:
		return nil, nil
	case []byte:
		if len(b) == 0 {
			return nil, nil
		}
		return b, nil
	case json.RawMessage:
		if len(b) == 0 {
			return nil, nil
		}
		return b, nil
	default:
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encoding request body: %w", err)
		}
		return raw, nil
	}
}

// PostMultipart sends a multipart/form-data POST (file upload) and returns the
// raw response body. Uploads are non-idempotent, so this does a single request
// (no auto-retry) but still applies auth and rate limiting. In dry-run it prints
// an equivalent curl and returns errDryRun.
func (c *Client) PostMultipart(ctx context.Context, path string, fields map[string]string, fileField, fileName string, fileData []byte) ([]byte, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		if v != "" {
			_ = mw.WriteField(k, v)
		}
	}
	fw, err := mw.CreateFormFile(fileField, fileName)
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(fileData); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	rawURL := c.buildURL(path, nil)
	if c.dryRun {
		c.printCurlMultipart(rawURL, fields, fileField, fileName)
		return nil, errDryRun
	}
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.apiKey != "" {
		req.Header.Set(apiKeyHeader, c.apiKey)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload to %s failed: %w", rawURL, err)
	}
	respBody, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("reading upload response: %w", readErr)
	}
	// React to the rate-limit budget like do() does, so uploads share the same
	// adaptive throttle/recovery rather than diverging.
	if resp.StatusCode == http.StatusTooManyRequests {
		c.rateLimiter.Throttle()
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, newAPIError(resp.StatusCode, respBody)
	}
	c.rateLimiter.Restore()
	return respBody, nil
}

// printCurlMultipart writes a copy-pasteable curl -F command for a dry-run upload.
func (c *Client) printCurlMultipart(rawURL string, fields map[string]string, fileField, fileName string) {
	var b strings.Builder
	fmt.Fprintf(&b, "curl -X POST %s", shellQuote(rawURL))
	key := "<redacted>"
	if c.showToken {
		key = c.apiKey
	}
	b.WriteString(" \\\n  -H " + shellQuote(apiKeyHeader+": "+key))
	b.WriteString(" \\\n  -F " + shellQuote(fileField+"=@"+fileName))
	for k, v := range fields {
		if v != "" {
			b.WriteString(" \\\n  -F " + shellQuote(k+"="+v))
		}
	}
	fmt.Fprintln(c.dryRunWriter, b.String())
}

// printCurl writes a copy-pasteable, header-redacted curl command for dry-run.
func (c *Client) printCurl(method, rawURL string, body []byte) {
	var b strings.Builder
	fmt.Fprintf(&b, "curl -X %s %s", method, shellQuote(rawURL))
	b.WriteString(" \\\n  -H " + shellQuote("Accept: application/json"))
	key := "<redacted>"
	if c.showToken {
		key = c.apiKey
	}
	b.WriteString(" \\\n  -H " + shellQuote(apiKeyHeader+": "+key))
	if body != nil {
		b.WriteString(" \\\n  -H " + shellQuote("Content-Type: application/json"))
		b.WriteString(" \\\n  -d " + shellQuote(string(body)))
	}
	fmt.Fprintln(c.dryRunWriter, b.String())
}

// shellQuote single-quotes s for safe shell pasting.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
