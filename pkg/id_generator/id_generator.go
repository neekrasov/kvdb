package idgenerator

import (
	"math"
	"sync/atomic"
)

// IDGenerator - struct for generating unique IDs using an atomic counter.
type IDGenerator struct {
	counter atomic.Int64
}

// NewIDGenerator - Initializes a new IDGenerator with a starting value.
func NewIDGenerator(previousID int64) *IDGenerator {
	generator := &IDGenerator{}
	generator.counter.Store(previousID)
	return generator
}

// Generate - Generates a new unique ID. Resets the counter if it reaches the maximum value.
func (g *IDGenerator) Generate() int64 {
	g.counter.CompareAndSwap(math.MaxInt64, 0)
	return g.counter.Add(1)
}
