package sync

import "sync"

// WithLock - acquires the given lock, executes the provided action.
func WithLock(l sync.Locker, action func()) {
	if action == nil {
		return
	}

	l.Lock()
	defer l.Unlock()
	action()
}
