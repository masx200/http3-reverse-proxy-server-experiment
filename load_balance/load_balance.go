package load_balance

import "net/http"

// LoadBalance 是一个负载均衡接口，它定义了如何对HTTP请求进行负载均衡转发。
// 其中包含了一个Map，用于映射域名到对应的UpStream。
type LoadBalance interface {
	// RoundTrip 是一个代理方法，用于发送HTTP请求，并返回响应或错误。
	RoundTrip(*http.Request) (*http.Response, error)
	// Map 是一个键值对映射，键是字符串类型，值是UpStream类型。
	Map[string, UpStream]
}

// UpStream 是一个上游服务接口，定义了如何与上游服务进行交互以及健康检查的方法。
type UpStream interface {
	// RoundTrip 是一个代理方法，用于发送HTTP请求，并返回响应或错误。
	RoundTrip(*http.Request) (*http.Response, error)
	// HealthyCheck 用于检查上游服务的健康状态，并返回健康状态及错误信息。
	HealthyCheck() (bool, error)
	// Identifier 用于返回上游服务的唯一标识符。
	Identifier() string
}

// Map 是一个泛型映射接口，支持基本的映射操作。
type Map[T comparable, Y any] interface {
	// Clear 清空映射中的所有元素。
	Clear()
	// Delete 从映射中删除指定的键。
	Delete(T)
	// Get 返回指定键的值，如果键不存在，则返回false。
	Get(T) (Y, bool)
	// Set 设置指定键的值。
	Set(T, Y)
	// Has 检查映射中是否存在指定的键。
	Has(T) bool
	// Values 返回映射中所有值的切片。
	Values() []Y
	// Kes 返回映射中所有键的切片。
	Kes() []T
	// Size 返回映射中元素的数量。
	Size() int64
	// Entries 返回映射中所有键值对的切片。
	Entries() []Pair[T, Y]
}

// Pair 是一个泛型结构体，用于存储一对任意类型的值。
// T和Y是泛型参数，代表First和Second可以是任何类型。
type Pair[T any, Y any] struct {
	First  T // First是结构体中的第一个元素。
	Second Y // Second是结构体中的第二个元素。
}
