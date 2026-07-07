package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type loginAttempt struct {
	count        int
	firstSeen    time.Time
	blockedUntil time.Time
}

type LoginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt

	maxAttempts   int
	window        time.Duration
	blockDuration time.Duration
}

// NewLoginRateLimiter creates an in-memory per-instance limiter.
// If the app runs multiple API replicas behind a load balancer, swap
// the map for Redis (INCR + EXPIRE) so counters are shared cluster-wide.
func NewLoginRateLimiter(maxAttempts int, window, blockDuration time.Duration) *LoginRateLimiter {
	rl := &LoginRateLimiter{
		attempts:      make(map[string]*loginAttempt),
		maxAttempts:   maxAttempts,
		window:        window,
		blockDuration: blockDuration,
	}
	go rl.cleanupLoop()
	return rl
}

func rateLimitKey(ip, email string) string {
	return ip + "|" + email
}

func (rl *LoginRateLimiter) Allow(ip, email string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := rateLimitKey(ip, email)
	a, ok := rl.attempts[key]
	if !ok {
		return true
	}

	now := time.Now()

	if now.Before(a.blockedUntil) {
		return false
	}

	if now.Sub(a.firstSeen) > rl.window {
		delete(rl.attempts, key)
		return true
	}

	return a.count < rl.maxAttempts
}

func (rl *LoginRateLimiter) RecordFailure(ip, email string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := rateLimitKey(ip, email)
	now := time.Now()

	a, ok := rl.attempts[key]
	if !ok || now.Sub(a.firstSeen) > rl.window {
		a = &loginAttempt{firstSeen: now}
		rl.attempts[key] = a
	}

	a.count++
	if a.count >= rl.maxAttempts {
		a.blockedUntil = now.Add(rl.blockDuration)
	}
}

func (rl *LoginRateLimiter) Reset(ip, email string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, rateLimitKey(ip, email))
}

func (rl *LoginRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, a := range rl.attempts {
			if now.Sub(a.firstSeen) > rl.window && now.After(a.blockedUntil) {
				delete(rl.attempts, key)
			}
		}
		rl.mu.Unlock()
	}
}

func ClientIP(r *http.Request) string {
	if fly := r.Header.Get("Fly-Client-IP"); fly != "" {
		return fly
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
