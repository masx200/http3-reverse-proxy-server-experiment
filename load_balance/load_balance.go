package load_balance

import "net/http"
import "github.com/moznion/go-optional"

// LoadBalance 是一个负载均衡接口，它定义了如何对HTTP请求进行负载均衡转发。
// 其中包含了一个Map，用于映射域名到对应的UpStream。
type LoadBalanceAndUpStream interface {
	// RoundTrip 是一个代理方法，用于发送HTTP请求，并返回响应或错误。
	// 参数：
	//   *http.Request: 待发送的HTTP请求
	// 返回值：
	//   *http.Response: 请求的响应
	//   error: 请求过程中遇到的错误
	RoundTrip(*http.Request) (*http.Response, error)

	// UpStreams 返回一个键值对映射，其中键是字符串类型，表示域名；
	// 值是UpStream类型，表示对应域名的上游服务集群。
	// 这个方法用于获取当前负载均衡器中配置的所有上游服务信息。
	UpStreams() optional.Option[MapInterface[string, LoadBalanceAndUpStream]]

	//选择一个可用的上游服务器
	// 参数：
	SelectAvailableServer() (LoadBalanceAndUpStream, error)

	// ActiveHealthyCheck 用于检查上游服务的健康状态，并返回健康状态及错误信息。
	// 返回值：bool - 上游服务的主动健康状态（true为健康，false为不健康）；error - 错误信息（如果有）
	ActiveHealthyCheck() (bool, error)

	// Identifier 用于返回上游服务的唯一标识符。
	// 返回值：string - 上游服务的唯一标识符
	Identifier() string

	// IsHealthy 用于判断上游服务是否健康。
	// 返回值：bool - 上游服务的健康状态（true为健康，false为不健康）
	IsHealthy() bool

	// SetHealthy 用于标记上游服务的健康状态。
	// 参数：bool - 上游服务的健康状态（true为健康，false为不健康）
	SetHealthy(bool)

	// IsHealthyResponse 根据HTTP响应判断上游服务是否健康。
	// 参数：*http.Response - 上游服务返回的HTTP响应
	// 返回值：bool - 上游服务的被动健康状态（true为健康，false为不健康）
	IsHealthyResponse(*http.Response) (bool, error)
}

// UpStream 是一个上游服务接口，定义了如何与上游服务进行交互以及健康检查的方法。
// 该接口包括发送HTTP请求、健康检查、标识服务和标记健康状态等方法。

// MapInterface 是一个泛型映射接口，支持基本的映射操作。
type MapInterface[T comparable, Y any] interface {
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
	Keys() []T
	// Size 返回映射中元素的数量。
	Size() int64
	// Entries 返回映射中所有键值对的切片。
	Entries() []PairInterface[T, Y]
}

// Pair 是一个泛型结构体，用于存储一对任意类型的值。
// T和Y是泛型参数，代表First和Second可以是任何类型。
type Pair[T any, Y any] struct {
	First  T // First是结构体中的第一个元素。
	Second Y // Second是结构体中的第二个元素。
}
type PairInterface[T any, Y any] interface {
	GetFirst() T
	SetFirst(T)
	GetSecond() Y

	SetSecond(Y)
}
