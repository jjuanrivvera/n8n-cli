package api

import (
	"context"
	"log/slog"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter is a client-side token-bucket limiter. The n8n public API does not
// document quota response headers (no X-RateLimit-*), so this is a fixed RPS
// budget that reacts to 429s: Throttle halves the rate, Restore gradually returns
// it to the configured base. This keeps a self-hosted instance from being hammered
// and stays well under n8n Cloud's limits.
type RateLimiter struct {
	mu      sync.Mutex
	limiter *rate.Limiter
	base    float64
	current float64
	logger  *slog.Logger
}

// NewRateLimiter creates a limiter at requestsPerSecond. A value <= 0 disables
// limiting (effectively unlimited), which is useful in tests.
func NewRateLimiter(requestsPerSecond float64, logger *slog.Logger) *RateLimiter {
	if logger == nil {
		logger = slog.Default()
	}
	if requestsPerSecond <= 0 {
		return &RateLimiter{limiter: rate.NewLimiter(rate.Inf, 0), base: 0, current: 0, logger: logger}
	}
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
		base:    requestsPerSecond,
		current: requestsPerSecond,
		logger:  logger,
	}
}

// Wait blocks until the limiter allows another request or ctx is done.
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	l := r.limiter
	r.mu.Unlock()
	return l.Wait(ctx)
}

// Throttle halves the current rate (floored at 0.5 rps) after a 429.
func (r *RateLimiter) Throttle() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.base == 0 { // unlimited mode, nothing to throttle
		return
	}
	r.current /= 2
	if r.current < 0.5 {
		r.current = 0.5
	}
	r.limiter.SetLimit(rate.Limit(r.current))
	r.logger.Debug("rate limited: throttling", "rps", r.current)
}

// Restore nudges the rate back toward the configured base after a success.
func (r *RateLimiter) Restore() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.base == 0 || r.current >= r.base {
		return
	}
	r.current += r.base * 0.1
	if r.current > r.base {
		r.current = r.base
	}
	r.limiter.SetLimit(rate.Limit(r.current))
}
