package sync

import (
	"math"
	"sync/atomic"
)

// IDGenerator - struct for generating unique IDs using an atomic counter.
type IDGenerator struct {
	counter atomic.Int64
}

// NewIDGenerator - Initializes a new IDGenerator with a starting value.
func NewIDGenerator(prevID int64) *IDGenerator {
	gen := &IDGenerator{}
	gen.counter.Store(prevID)
	return gen
}

// Generate - Generates a new unique ID. Resets the counter if it reaches the maximum value.
func (g *IDGenerator) Generate() int64 {
	g.counter.CompareAndSwap(math.MaxInt64, 0)
	return g.counter.Add(1)
}
