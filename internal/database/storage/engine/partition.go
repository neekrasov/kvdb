package engine

import (
	"context"
	"sync"
	"time"

	pkgsync "github.com/neekrasov/kvdb/pkg/sync"
)

// value stores the value and its expiration time.
type value struct {
	Value string
	TTL   int64
}

// partitionMap - represents one data partition.
type partitionMap struct {
	mu       sync.RWMutex
	data     map[string]value
	watchers map[string]*watcher
}

// newPartMap - returns a new partition instance
func newPartMap() *partitionMap {
	return &partitionMap{
		data:     make(map[string]value),
		watchers: map[string]*watcher{},
	}
}

// set - set stores a key-value pair in memory.
func (p *partitionMap) set(key, val string, ttl int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	watcher, ok := p.watchers[key]
	if ok {
		watcher.set(val)
	}

	p.data[key] = value{Value: val, TTL: ttl}
}

// get - retrieves the value associated with a key.
func (p *partitionMap) get(key string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	val, exists := p.data[key]

	if val.TTL > 0 && time.Now().Unix() > val.TTL {
		delete(p.data, key)
		return "", false
	}

	return val.Value, exists
}

// del - removes a key-value pair from memory.
func (p *partitionMap) del(key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.data, key)
	return nil
}

// watch - watches the key and returns the value if it has changed.
func (p *partitionMap) watch(ctx context.Context, key string) pkgsync.FutureString {
	p.mu.Lock()
	defer p.mu.Unlock()

	future := pkgsync.NewFuture[string]()
	go func() {
		watcher, ok := p.watchers[key]
		if ok {
			future.Set(watcher.watch(ctx))
		}

		val, ok := p.get(key)
		if !ok {
			val = ""
		}

		watcher = newWatcher(val)
		p.watchers[key] = watcher
		future.Set(watcher.watch(ctx))
	}()

	return future
}
