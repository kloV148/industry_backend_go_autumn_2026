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

	var last time.Time

	if clock != nil {
		last = clock.Now()
	}

	return &Limiter{
		rate:   ratePerSec,
		burst:  burst,
		tokens: float64(burst),
		clock:  clock,
		last:   last,
	}
}

func (l *Limiter) AllowN(n int) bool {
	if l.burst <= 0 || l.clock == nil || n <= 0 {
		return false
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock.Now()
	elapsed := now.Sub(l.last).Seconds()

	if elapsed > 0 {
		l.last = now

		if l.rate > 0 {
			l.tokens += elapsed * l.rate
		}

		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}
	}

	if l.tokens >= float64(n) {
		l.tokens -= float64(n)
		return true
	}

	return false
}
