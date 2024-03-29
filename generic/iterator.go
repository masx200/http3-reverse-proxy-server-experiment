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
