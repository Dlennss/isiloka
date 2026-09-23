package helper

import (
	"net/http"
	"sync"
	"time"
)

type ipEntry struct {
	count    int
	windowAt time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipEntry
	max     int
	window  time.Duration
}

func NewRateLimiter(maxPerWindow int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*ipEntry),
		max:     maxPerWindow,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.window * 2)
		rl.mu.Lock()
		now := time.Now()
		for k, v := range rl.entries {
			if now.Sub(v.windowAt) > rl.window*2 {
				delete(rl.entries, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	e, ok := rl.entries[ip]
	if !ok || now.Sub(e.windowAt) > rl.window {
		rl.entries[ip] = &ipEntry{count: 1, windowAt: now}
		return true
	}
	e.count++
	return e.count <= rl.max
}

func (rl *RateLimiter) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := extractClientIP(r)
		if !rl.Allow(ip) {
			http.Error(w, `{"ok":false,"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
