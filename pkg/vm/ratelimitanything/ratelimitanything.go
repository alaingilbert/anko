package ratelimitanything

import (
	"context"
	"sync"
	"time"
)

// RateLimitAnything ...
type RateLimitAnything struct {
	sync.Mutex
	counter                   int64
	limit                     int64
	start                     time.Time
	period                    time.Duration
	RateLimitExceededCallback func(duration time.Duration)
}

// NewRateLimitAnything ...
func NewRateLimitAnything(limit int64, period time.Duration) *RateLimitAnything {
	r := new(RateLimitAnything)
	r.limit = limit
	r.period = period
	r.RateLimitExceededCallback = func(duration time.Duration) {}
	return r
}

// Get ...
func (r *RateLimitAnything) Get() <-chan struct{} {
	return r.get(context.Background())
}

// GetWithContext ...
func (r *RateLimitAnything) GetWithContext(ctx context.Context) <-chan struct{} {
	return r.get(ctx)
}

func (r *RateLimitAnything) get(ctx context.Context) <-chan struct{} {
	r.Lock()
	defer r.Unlock()
	now := time.Now()
	end := r.start.Add(r.period)
	if now.After(end) {
		r.start = now
		r.counter = 0
	}
	r.counter++
	ch := make(chan struct{})
	if r.counter <= r.limit {
		close(ch)
	} else {
		remaining := end.Sub(now)
		r.RateLimitExceededCallback(remaining)
		go func() {
			select {
			case <-time.After(remaining):
			case <-ctx.Done():
			}
			close(ch)
		}()
	}
	return ch
}
