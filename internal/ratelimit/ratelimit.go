// Package ratelimit caps how often one key may do something.
//
// It exists for one reason: guessing a password is the cheapest attack on this
// platform, and until now nothing stopped anyone trying.
//
// The counting is in memory, which means per process. The API runs at most
// three instances, so an attacker spread across them gets at most three times
// the limit — tolerable for a platform at this size, and the number is stated
// rather than hidden. Moving the counters to the database is the fix when
// that stops being true.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter allows at most n attempts per key within a sliding window.
type Limiter struct {
	limit  int
	window time.Duration

	mu   sync.Mutex
	seen map[string][]time.Time

	// now is injectable so the window can be tested without sleeping.
	now func() time.Time
}

// New builds a limiter.
func New(limit int, window time.Duration) *Limiter {
	return &Limiter{
		limit:  limit,
		window: window,
		seen:   map[string][]time.Time{},
		now:    time.Now,
	}
}

// Allow records an attempt and reports whether it is permitted. When it is
// not, it returns how long until the oldest attempt ages out — a refusal that
// does not say when to come back tells the person nothing they can act on.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	cutoff := now.Add(-l.window)

	// Sweeping every call keeps this honest about memory without a goroutine
	// to supervise: keys arrive from the internet — an email in a login body
	// is attacker controlled — and a map that only grows is the memory of the
	// process.
	l.sweep(cutoff)

	attempts := keep(l.seen[key], cutoff)
	if len(attempts) >= l.limit {
		l.seen[key] = attempts
		return false, attempts[0].Add(l.window).Sub(now)
	}

	l.seen[key] = append(attempts, now)
	return true, 0
}

// sweep drops keys whose attempts have all aged out.
func (l *Limiter) sweep(cutoff time.Time) {
	for key, attempts := range l.seen {
		if left := keep(attempts, cutoff); len(left) == 0 {
			delete(l.seen, key)
		} else {
			l.seen[key] = left
		}
	}
}

func keep(attempts []time.Time, cutoff time.Time) []time.Time {
	i := 0
	for i < len(attempts) && !attempts[i].After(cutoff) {
		i++
	}
	return attempts[i:]
}

// size is for tests, which have to be able to assert that memory is reclaimed.
func (l *Limiter) size() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.seen)
}

// Peek reports whether the key is currently inside its limit, without
// recording an attempt.
//
// Separate from Allow because a successful sign-in must not count against the
// person: the check happens before the password is verified and the record
// happens only if it was wrong.
func (l *Limiter) Peek(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	attempts := keep(l.seen[key], now.Add(-l.window))
	if len(attempts) >= l.limit {
		return false, attempts[0].Add(l.window).Sub(now)
	}
	return true, 0
}
