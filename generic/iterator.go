package generic

type Iterable[T any] interface {
	Iterator() Iterator[T]
}
type Iterator[T any] interface {
	Next() (T, bool)
}
