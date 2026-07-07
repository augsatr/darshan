package db

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestViewBuffer_NoDropUnderConcurrentLoad(t *testing.T) {
	var mu sync.Mutex
	var flushed []int64

	vb := &ViewBuffer{
		buf:      make([]int64, 0, 100),
		maxSize:  100,
		interval: time.Hour, // never flushes via ticker
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		flushFn: func(_ context.Context, batch []int64) error {
			mu.Lock()
			flushed = append(flushed, batch...)
			mu.Unlock()
			return nil
		},
	}

	go vb.loop()

	var wg sync.WaitGroup
	n := 1000
	for i := 0; i < n; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			vb.Add(int64(i))
		}()
	}
	wg.Wait()

	vb.Stop()

	mu.Lock()
	got := len(flushed)
	mu.Unlock()

	if got != n {
		t.Fatalf("expected %d flushed entries, got %d", n, got)
	}

	// Verify buffer didn't drop or duplicate entries under concurrent Add/Flush.
	// Multiple views of the same temple_id are valid in page_views; this check
	// only confirms the buffer doesn't corrupt its own internal state.
	seen := make(map[int64]bool)
	mu.Lock()
	for _, id := range flushed {
		if seen[id] {
			t.Fatalf("duplicate entry: %d", id)
		}
		seen[id] = true
	}
	mu.Unlock()

	if len(seen) != n {
		t.Fatalf("expected %d unique IDs, got %d", n, len(seen))
	}
}

func TestViewBuffer_FlushesOnCapacity(t *testing.T) {
	var mu sync.Mutex
	var flushCount atomic.Int32

	vb := &ViewBuffer{
		buf:      make([]int64, 0, 5),
		maxSize:  5,
		interval: time.Hour,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		flushFn: func(_ context.Context, batch []int64) error {
			flushCount.Add(1)
			mu.Lock()
			if len(batch) > 5 {
				t.Errorf("batch too large: %d", len(batch))
			}
			mu.Unlock()
			return nil
		},
	}

	go vb.loop()

	for i := 0; i < 10; i++ {
		vb.Add(int64(i))
	}

	vb.Stop()

	if c := flushCount.Load(); c < 2 {
		t.Fatalf("expected at least 2 flushes (capacity 5, 10 items), got %d", c)
	}
}

func TestViewBuffer_EmptyStopDoesntCrash(t *testing.T) {
	vb := &ViewBuffer{
		buf:      make([]int64, 0, 10),
		maxSize:  10,
		interval: time.Hour,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		flushFn: func(_ context.Context, _ []int64) error {
			return nil
		},
	}

	go vb.loop()

	// Add nothing, stop immediately
	vb.Stop()
}

func TestViewBuffer_TickerFlushes(t *testing.T) {
	var flushed int

	vb := &ViewBuffer{
		buf:      make([]int64, 0, 100),
		maxSize:  100, // higher than items added, so only ticker triggers flush
		interval: 10 * time.Millisecond,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		flushFn: func(_ context.Context, batch []int64) error {
			flushed += len(batch)
			return nil
		},
	}

	go vb.loop()

	vb.Add(1)
	vb.Add(2)
	vb.Add(3)

	time.Sleep(50 * time.Millisecond)
	vb.Stop()

	if flushed != 3 {
		t.Fatalf("expected 3 flushed, got %d", len(vb.buf))
	}
}

func TestViewBuffer_NoDoubleFlushRace(t *testing.T) {
	var mu sync.Mutex
	var flushCount int

	vb := &ViewBuffer{
		buf:      make([]int64, 0, 3),   // small: flushes every 3
		maxSize:  3,
		interval: 10 * time.Millisecond, // aggressive ticker
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		flushFn: func(_ context.Context, batch []int64) error {
			mu.Lock()
			flushCount++
			mu.Unlock()
			_ = batch
			time.Sleep(5 * time.Millisecond) // slow flush to increase race window
			return nil
		},
	}

	go vb.loop()

	for i := 0; i < 9; i++ {
		vb.Add(int64(i))
		time.Sleep(2 * time.Millisecond) // overlap with ticker
	}

	time.Sleep(50 * time.Millisecond)
	vb.Stop()

	mu.Lock()
	fc := flushCount
	mu.Unlock()
	if fc == 0 {
		t.Fatal("expected at least one flush")
	}
}
