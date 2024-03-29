package load_balance

import (
	"net/http"

	"github.com/masx200/http3-reverse-proxy-server-experiment/generic"
	"github.com/moznion/go-optional"
)

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
	GetUpStreams() optional.Option[generic.MapInterface[string, LoadBalanceAndUpStream]]

	//选择一个可用的上游服务器
	// 参数：
	SelectAvailableServer() (LoadBalanceAndUpStream, error)

	// ActiveHealthyCheck 用于检查上游服务的健康状态，并返回健康状态及错误信息。
	// 返回值：bool - 上游服务的主动健康状态（true为健康，false为不健康）；error - 错误信息（如果有）
	ActiveHealthyCheck() (bool, error)

	// Identifier 用于返回上游服务的唯一标识符。
	// 返回值：string - 上游服务的唯一标识符
	GetIdentifier() string

	// IsHealthy 用于判断上游服务是否健康。
	// 返回值：bool - 上游服务的健康状态（true为健康，false为不健康）
	GetHealthy() bool

	// SetHealthy 用于标记上游服务的健康状态。
	// 参数：bool - 上游服务的健康状态（true为健康，false为不健康）
	SetHealthy(bool)

	// PassiveUnHealthyCheck 根据HTTP响应判断上游服务是否健康。
	// 参数：*http.Response - 上游服务返回的HTTP响应
	// 返回值：bool - 上游服务的被动健康状态（true为健康，false为不健康）
	PassiveUnHealthyCheck(*http.Response) (bool, error)

	// SetHealthCheckInterval 设置健康状态的缓存最大年龄。
	// 参数:
	//   int64 - 表示健康状态的缓存的最大年龄（单位：毫秒）。
	SetHealthyCheckInterval(int64)
	// 设置逻辑实现

	// GetHealthCheckInterval 获取健康状态的缓存最大年龄。
	// 返回值:
	//   int64 - 健康状态的缓存的最大年龄（单位：毫秒）。
	GetHealthyCheckInterval() int64
	// 获取逻辑实现
	SetUnHealthyFailDuration(int64)
	// 设置逻辑实现

	// GetHealthCheckInterval 获取健康状态的缓存最大年龄。
	// 返回值:
	//   int64 - 健康状态的缓存的最大年龄（单位：毫秒）。
	GetUnHealthyFailDuration() int64
}

// UpStream 是一个上游服务接口，定义了如何与上游服务进行交互以及健康检查的方法。
// 该接口包括发送HTTP请求、健康检查、标识服务和标记健康状态等方法。
