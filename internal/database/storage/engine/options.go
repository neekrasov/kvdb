package engine

// Option - options for configuring Engine.
type Option func(*Engine)

// WithPartitionNum - configures Engine with a partition number.
func WithPartitionNum(partnum int) Option {
	return func(e *Engine) {
		e.partitions = make([]*partitionMap, partnum)
		for i := 0; i < int(partnum); i++ {
			e.partitions[i] = newPartMap()
		}
	}
}
