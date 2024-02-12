package utils

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"sync"
	"time"
)

func Ternary[T any](predicate bool, a, b T) T {
	if predicate {
		return a
	}
	return b
}

// MD5 returns md5 hex sum as a string
func MD5(in []byte) string {
	h := md5.New()
	h.Write(in)
	return hex.EncodeToString(h.Sum(nil))
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func MustErr[T any](v T, err error) error {
	if err == nil {
		panic("error expected")
	}
	return err
}

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
