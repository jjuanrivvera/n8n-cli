package api

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// backoff must not panic when randFloat is nil (the default-0.5 path).
func TestBackoff_NilRandFloat(t *testing.T) {
	p := &RetryPolicy{InitialBackoff: time.Second, MaxBackoff: 10 * time.Second} // randFloat nil
	d := p.backoff(0, nil)
	assert.InDelta(t, float64(500*time.Millisecond), float64(d), float64(time.Millisecond))
}

func TestRetryAfter_HTTPDate(t *testing.T) {
	future := time.Now().Add(90 * time.Second).UTC().Format(http.TimeFormat)
	d, ok := retryAfter(future)
	require.True(t, ok)
	assert.Greater(t, d, 60*time.Second)

	past := time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)
	d, ok = retryAfter(past)
	require.True(t, ok)
	assert.Equal(t, time.Duration(0), d) // clamped

	_, ok = retryAfter("not a date or number")
	assert.False(t, ok)
}

func TestRateLimiter_Math(t *testing.T) {
	rl := NewRateLimiter(10, slog.Default())
	assert.InDelta(t, 10.0, float64(rl.limiter.Limit()), 0.001)
	rl.Throttle()
	assert.InDelta(t, 5.0, float64(rl.limiter.Limit()), 0.001)
	// throttle to the floor
	for range 10 {
		rl.Throttle()
	}
	assert.GreaterOrEqual(t, float64(rl.limiter.Limit()), 0.5)
	// restore climbs back toward base and caps there
	for range 100 {
		rl.Restore()
	}
	assert.InDelta(t, 10.0, float64(rl.limiter.Limit()), 0.001)
}

func TestAPIError_403_429_Hints(t *testing.T) {
	e403 := newAPIError(403, []byte(`{"message":"nope"}`))
	assert.True(t, e403.IsForbidden())
	assert.Contains(t, e403.Error(), "scope")

	e429 := newAPIError(429, []byte(`{"message":"slow down"}`))
	assert.True(t, e429.IsRateLimited())
	assert.Contains(t, e429.Error(), "rate limited")

	// description-only body becomes the message
	eDesc := newAPIError(400, []byte(`{"description":"only this"}`))
	assert.Equal(t, "only this", eDesc.Message)
}

func TestDo_ContextCancelDuringBackoff(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500) // always retryable -> client enters backoff
	})
	c.retryPolicy.MaxRetries = 5
	c.retryPolicy.InitialBackoff = 10 * time.Second
	c.retryPolicy.MaxBackoff = 30 * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(50 * time.Millisecond); cancel() }() // cancel mid-backoff
	start := time.Now()
	err := c.Do(ctx, "GET", "/workflows", nil, nil, nil)
	require.Error(t, err)
	assert.Less(t, time.Since(start), 5*time.Second) // did not wait the full 10s backoff
}
