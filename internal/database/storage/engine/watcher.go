package engine

import (
	"context"
	"sync"
)

// watcher - a structure representing a watcher for value changes.
type watcher struct {
	mu    *sync.Mutex
	cond  *sync.Cond
	value string
}

// newWatcher - creates a new instance of watcher with the specified initial value.
func newWatcher(value string) *watcher {
	mu := sync.Mutex{}
	cond := sync.NewCond(&mu)

	return &watcher{
		mu:    &mu,
		cond:  cond,
		value: value,
	}
}

// set - updates the value and notifies all waiting watchers.
func (w *watcher) set(value string) {
	w.mu.Lock()
	w.value = value
	w.mu.Unlock()

	w.cond.Broadcast()
}

// watch - waits for a value change and returns the new value when it changes or when the context is canceled.
func (w *watcher) watch(ctx context.Context) string {
	w.mu.Lock()
	defer w.mu.Unlock()

	oldVal := w.value
	for oldVal == w.value {
		select {
		case <-ctx.Done():
			return w.value
		default:
			w.cond.Wait()
		}
	}

	return w.value
}
