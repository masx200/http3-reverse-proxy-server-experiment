package generic

// 构造函数
func NewPairImplement[T any, Y any](first T, second Y) PairInterface[T, Y] {
	return &PairImplement[T, Y]{
		First:  first,
		Second: second,
	}
}

// Pair 是一个泛型结构体，用于存储一对任意类型的值。
// T和Y是泛型参数，代表First和Second可以是任何类型。
type PairImplement[T any, Y any] struct {
	First  T // First是结构体中的第一个元素。
	Second Y // Second是结构体中的第二个元素。
}
type PairInterface[T any, Y any] interface {
	GetFirst() T
	SetFirst(T)
	GetSecond() Y

	SetSecond(Y)
}

// 实现 GetFirst 方法
func (p *PairImplement[T, Y]) GetFirst() T {
	return p.First
}

// 实现 SetFirst 方法
func (p *PairImplement[T, Y]) SetFirst(first T) {
	p.First = first
}

// 实现 GetSecond 方法
func (p *PairImplement[T, Y]) GetSecond() Y {
	return p.Second
}

// 实现 SetSecond 方法
func (p *PairImplement[T, Y]) SetSecond(second Y) {
	p.Second = second
}
