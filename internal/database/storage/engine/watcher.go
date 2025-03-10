package engine

import (
	"context"
	"sync"
)

type Watcher struct {
	mu    *sync.Mutex
	cond  *sync.Cond
	value string
}

func NewWatcher(value string) *Watcher {
	mu := sync.Mutex{}
	cond := sync.NewCond(&mu)

	return &Watcher{
		mu:    &mu,
		cond:  cond,
		value: value,
	}
}

func (w *Watcher) Set(value string) {
	w.mu.Lock()
	w.value = value
	w.mu.Unlock()

	w.cond.Broadcast()
}

func (w *Watcher) Watch(ctx context.Context) string {
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
