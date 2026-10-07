package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	limiter := &Limiter{clock: clock, rate: ratePerSec, burst: burst}

	if clock != nil && burst > 0 {
		limiter.last = clock.Now()
		limiter.tokens = float64(burst)
	}

	return limiter
}
func (l *Limiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if n <= 0 || n > l.burst {
		return false
	}

	if l.burst <= 0 || l.clock == nil {
		return false
	}

	if l.rate > 0 {
		now := l.clock.Now()
		if now.After(l.last) {
			temp := now.Sub(l.last).Seconds()
			l.tokens = l.tokens + l.rate*temp
			l.last = now
		}
		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}
	}

	if l.tokens < float64(n) {
		return false
	}
	l.tokens = l.tokens - float64(n)
	return true

}
