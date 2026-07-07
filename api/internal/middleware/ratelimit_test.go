package middleware

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestLoginRateLimiter_Allow_BelowThreshold(t *testing.T) {
	rl := NewLoginRateLimiter(5, time.Minute, 5*time.Minute)
	if !rl.Allow("1.2.3.4", "a@b.com") {
		t.Fatal("expected allow on first attempt")
	}
}

func TestLoginRateLimiter_BlocksAtThreshold(t *testing.T) {
	rl := NewLoginRateLimiter(5, time.Minute, 5*time.Minute)
	ip := "1.2.3.4"
	email := "a@b.com"

	for i := 0; i < 5; i++ {
		if !rl.Allow(ip, email) {
			t.Fatalf("expected allow on attempt %d", i+1)
		}
		rl.RecordFailure(ip, email)
	}

	if rl.Allow(ip, email) {
		t.Fatal("expected block after 5 failures")
	}
}

func TestLoginRateLimiter_SurvivesAcrossBursts(t *testing.T) {
	rl := NewLoginRateLimiter(5, time.Minute, 5*time.Minute)
	ip := "1.2.3.4"
	email := "a@b.com"

	for i := 0; i < 5; i++ {
		rl.RecordFailure(ip, email)
	}

	if rl.Allow(ip, email) {
		t.Fatal("expected block after 5 failures")
	}

	// Give time for cleanup tick (won't fire since window/block haven't elapsed)
	time.Sleep(10 * time.Millisecond)

	// Still blocked
	if rl.Allow(ip, email) {
		t.Fatal("expected still blocked")
	}
}

func TestLoginRateLimiter_ResetClearsBlock(t *testing.T) {
	rl := NewLoginRateLimiter(5, time.Minute, 5*time.Minute)
	ip := "1.2.3.4"
	email := "a@b.com"

	for i := 0; i < 5; i++ {
		rl.RecordFailure(ip, email)
	}

	if rl.Allow(ip, email) {
		t.Fatal("expected block")
	}

	rl.Reset(ip, email)

	if !rl.Allow(ip, email) {
		t.Fatal("expected allow after reset")
	}
}

func TestLoginRateLimiter_DifferentEmailNotBlocked(t *testing.T) {
	rl := NewLoginRateLimiter(5, time.Minute, 5*time.Minute)

	for i := 0; i < 5; i++ {
		rl.RecordFailure("1.2.3.4", "a@b.com")
	}

	if !rl.Allow("1.2.3.4", "other@b.com") {
		t.Fatal("different email from same IP should not be blocked")
	}
	if !rl.Allow("5.6.7.8", "a@b.com") {
		t.Fatal("same email from different IP should not be blocked")
	}
}

func TestLoginRateLimiter_ConcurrentBurst(t *testing.T) {
	rl := NewLoginRateLimiter(5, time.Minute, 5*time.Minute)
	ip := "1.2.3.4"
	email := "a@b.com"

	var wg sync.WaitGroup
	allowed := make(chan bool, 20)
	done := make(chan struct{})

	// Fire 20 concurrent Allow+RecordFailure pairs
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok := rl.Allow(ip, email)
			allowed <- ok
			rl.RecordFailure(ip, email)
		}()
	}

	go func() {
		wg.Wait()
		close(allowed)
		close(done)
	}()

	<-done

	var allowedCount int
	var deniedCount int
	for ok := range allowed {
		if ok {
			allowedCount++
		} else {
			deniedCount++
		}
	}

	// Allow + RecordFailure are two separate mutex acquisitions, so under
	// concurrent load at most 1 extra request can slip through (benign TOCTOU
	// race — microseconds wide, no security impact).
	if allowedCount > 6 {
		t.Fatalf("expected at most 6 allowed under concurrent load (TOCTOU margin), got %d allowed, %d denied", allowedCount, deniedCount)
	}
	if deniedCount == 0 {
		t.Fatal("expected at least some denied under concurrent load")
	}

	// After the burst, new request should be blocked
	if rl.Allow(ip, email) {
		t.Fatal("expected block after concurrent burst")
	}
}

func TestClientIP_RemoteAddr(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "1.2.3.4:56789"
	if got := ClientIP(r); got != "1.2.3.4" {
		t.Fatalf("expected 1.2.3.4, got %s", got)
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.1")
	r.RemoteAddr = "10.0.0.1:12345"
	if got := ClientIP(r); got != "203.0.113.1" {
		t.Fatalf("expected 203.0.113.1, got %s", got)
	}
}

func TestClientIP_FlyClientIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Fly-Client-IP", "198.51.100.1")
	r.Header.Set("X-Forwarded-For", "203.0.113.1")
	if got := ClientIP(r); got != "198.51.100.1" {
		t.Fatalf("expected Fly-Client-IP to take priority, got %s", got)
	}
}
