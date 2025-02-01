package sync

// Semaphore represents a simple semaphore to control concurrency.
type Semaphore struct {
	limit uint          // Maximum number of concurrent acquisitions.
	sem   chan struct{} // Channel used to track available permits.
}

// NewSemaphore creates a new Semaphore with the specified limit.
func NewSemaphore(limit uint) *Semaphore {
	return &Semaphore{
		limit: limit,
		sem:   make(chan struct{}, limit),
	}
}

// Acquire acquires a permit from the semaphore, blocking if no permits are available.
func (s *Semaphore) Acquire() {
	if s == nil || s.sem == nil {
		return
	}

	s.sem <- struct{}{}
}

// Release releases a permit, allowing another goroutine to acquire it.
func (s *Semaphore) Release() {
	if s == nil || s.sem == nil {
		return
	}

	<-s.sem
}
