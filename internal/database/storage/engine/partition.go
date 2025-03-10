package engine

import (
	"context"
	"sync"

	pkgsync "github.com/neekrasov/kvdb/pkg/sync"
)

// partitionMap - represents one data partition.
type partitionMap struct {
	mu       sync.RWMutex
	data     map[string]string
	watchers map[string]*Watcher
}

// newPartMap - returns a new partition instance
func newPartMap() *partitionMap {
	return &partitionMap{
		data:     make(map[string]string),
		watchers: map[string]*Watcher{},
	}
}

// Set - set stores a key-value pair in memory.
func (p *partitionMap) Set(key, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	watcher, ok := p.watchers[key]
	if ok {
		watcher.Set(value)
	}

	p.data[key] = value
}

// Get - retrieves the value associated with a key.
func (p *partitionMap) Get(key string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	val, exists := p.data[key]
	return val, exists
}

// Del - removes a key-value pair from memory.
func (p *partitionMap) Del(key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.data, key)
	return nil
}

// Watch - watches the key and returns the value if it has changed.
func (p *partitionMap) Watch(ctx context.Context, key string) pkgsync.FutureString {
	p.mu.Lock()
	defer p.mu.Unlock()

	future := pkgsync.NewFuture[string]()
	go func() {
		watcher, ok := p.watchers[key]
		if ok {
			future.Set(watcher.Watch(ctx))
		}

		val, ok := p.Get(key)
		if !ok {
			val = ""
		}

		watcher = NewWatcher(val)
		p.watchers[key] = watcher
		future.Set(watcher.Watch(ctx))
	}()

	return future
}
