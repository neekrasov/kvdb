package sync_test

import (
	"sync"
	"testing"
	"time"

	pkgsync "github.com/neekrasov/kvdb/pkg/sync"
)

func TestSemaphore_AcquireRelease(t *testing.T) {
	t.Parallel()

	sem := pkgsync.NewSemaphore(2)

	sem.Acquire()
	sem.Acquire()

	done := make(chan struct{})
	go func() {
		sem.Acquire()
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("expected semaphore to block, but it acquired more permits")
	case <-time.After(100 * time.Millisecond):
		// expected behavior
	}

	sem.Release()
	sem.Release()
}

func TestSemaphore_Concurrency(t *testing.T) {
	t.Parallel()

	sem := pkgsync.NewSemaphore(3)
	var wg sync.WaitGroup
	var mu sync.Mutex
	counter := 0
	totalGoroutines := 10
	wg.Add(totalGoroutines)

	for i := 0; i < totalGoroutines; i++ {
		go func() {
			sem.Acquire()
			mu.Lock()
			counter++
			mu.Unlock()
			sem.Release()
			wg.Done()
		}()
	}

	wg.Wait()

	if counter != totalGoroutines {
		t.Fatalf("expected counter to be %d, got %d", totalGoroutines, counter)
	}
}

func TestSemaphore_NilSafety(t *testing.T) {
	t.Parallel()

	var sem *pkgsync.Semaphore
	sem.Acquire()
	sem.Release()
}
