package api

import (
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// RetryPolicy controls exponential backoff with jitter. Only idempotent methods
// (GET/HEAD/PUT/DELETE) are retried on transient failures; POST/PATCH are never
// auto-retried so we don't accidentally create or mutate twice. 429 is retried
// for every method because no work was performed.
type RetryPolicy struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Logger         *slog.Logger

	// rng is seeded per policy so jitter is testable/deterministic when needed.
	randFloat func() float64
}

// DefaultRetryPolicy returns sensible defaults: up to 4 retries, 500ms→8s backoff.
func DefaultRetryPolicy(logger *slog.Logger) *RetryPolicy {
	//nolint:gosec // jitter does not require a cryptographic RNG
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &RetryPolicy{
		MaxRetries:     4,
		InitialBackoff: 500 * time.Millisecond,
		MaxBackoff:     8 * time.Second,
		Logger:         logger,
		randFloat:      r.Float64,
	}
}

// idempotent reports whether a method is safe to auto-retry.
func idempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}

// shouldRetry decides whether to retry given the method, response, and transport error.
func (p *RetryPolicy) shouldRetry(method string, resp *http.Response, err error) bool {
	if err != nil {
		// Transport-level errors (timeouts, resets) are retried only for idempotent methods.
		return idempotent(method)
	}
	if resp == nil {
		return false
	}
	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return true // no work performed; safe regardless of method
	case resp.StatusCode >= 500:
		return idempotent(method)
	default:
		return false
	}
}

// backoff returns the delay before the next attempt (0-indexed). It honors a
// Retry-After header when present, otherwise uses exponential backoff with full
// jitter, capped at MaxBackoff.
func (p *RetryPolicy) backoff(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if d, ok := retryAfter(resp.Header.Get("Retry-After")); ok {
			return d
		}
	}
	base := float64(p.InitialBackoff) * math.Pow(2, float64(attempt))
	if base > float64(p.MaxBackoff) {
		base = float64(p.MaxBackoff)
	}
	jitter := p.randFloat()
	if p.randFloat == nil {
		jitter = 0.5
	}
	return time.Duration(base * jitter)
}

// retryAfter parses an RFC 9110 Retry-After value (delta-seconds or HTTP-date).
func retryAfter(v string) (time.Duration, bool) {
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(max(secs, 0)) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		return max(time.Until(t), 0), true
	}
	return 0, false
}
