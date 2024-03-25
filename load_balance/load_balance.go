package load_balance

import "net/http"

type LoadBalance interface {
	RoundTrip(*http.Request) (*http.Response, error)
	Map[string, UpStream]
}
type UpStream interface {
	RoundTrip(*http.Request) (*http.Response, error)
	HealthyCheck() (bool, error)
	Identifier() string
}
type Map[T comparable, Y any] interface {
	Clear()
	Delete(T)
	Get(T) (Y, bool)
	Set(T, Y)
	Has(T) bool
	Values() []Y
	Kes() []T
	Size() int64
	Entries() []Pair[T, Y]
}

// Pair是一个泛型结构体，用于存储一对任意类型的值。
// T和Y是泛型参数，代表First和Second可以是任何类型。
type Pair[T any, Y any] struct {
	First  T // First是结构体中的第一个元素。
	Second Y // Second是结构体中的第二个元素。
}
