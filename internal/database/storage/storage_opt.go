package storage

// StorageOpt - Options for configuring Storage.
type StorageOpt func(*Storage)

// WithWALOpt - Configures Storage with a WAL.
func WithWALOpt(w WAL) StorageOpt {
	return func(s *Storage) {
		s.wal = w
	}
}
