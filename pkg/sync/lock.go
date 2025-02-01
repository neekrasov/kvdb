package sync

import "sync"

func WithLock(l sync.Locker, action func()) {
	if action == nil {
		return
	}

	l.Lock()
	defer l.Unlock()
	action()
}
