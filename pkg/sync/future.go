package sync

type (
	FutureError  = Future[error]
	FutureString = Future[string]
)

type Future[T any] struct {
	result chan T
	used   bool
}

func NewFuture[T any]() Future[T] {
	return Future[T]{result: make(chan T)}
}

func (f *Future[T]) Get() T {
	return <-f.result
}

func (f *Future[T]) Set(value T) {
	if f.used {
		return
	}

	f.used = true
	f.result <- value
	close(f.result)
}
