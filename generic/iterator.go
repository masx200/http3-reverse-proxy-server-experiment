package generic

type Iterable[T any] interface {
	Iterator() Iterator[T]
}
type Iterator[T any] interface {
	Next() IteratorResult[T]
}
type IteratorResult[T any] interface {
	GetDone() bool
	GetValue() T
}
type IteratorResultImplement[T any] struct {
	Value T
	Done  bool
}

// GetDone implements IteratorResult.
func (i *IteratorResultImplement[T]) GetDone() bool {
	return i.Done
}

// GetValue implements IteratorResult.
func (i *IteratorResultImplement[T]) GetValue() T {
	return i.Value
}
