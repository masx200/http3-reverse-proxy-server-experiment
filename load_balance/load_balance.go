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

	// ActiveHealthyCheck 用于检查上游服务的健康状态，并返回健康状态及错误信息。
	// 返回值：bool - 上游服务的主动健康状态（true为健康，false为不健康）；error - 错误信息（如果有）

	GetLoadBalanceService() optional.Option[LoadBalanceService]

	GetServerConfigCommon() ServerConfigCommon
}

// ServerConfigCommon 定义了服务配置的公共接口
type ServerConfigCommon interface {
	GetPassiveUnHealthyCheckStatusCodeRange() generic.PairInterface[int, int]
	// GetActiveHealthyCheckURL 返回主动健康检查的URL
	GetActiveHealthyCheckURL() string
	// GetActiveHealthyCheckMethod 返回主动健康检查的方法（如GET、POST）
	GetActiveHealthyCheckMethod() string
	// GetActiveHealthyCheckStatusCodeRange 返回主动健康检查接受的状态码范围
	GetActiveHealthyCheckStatusCodeRange() generic.PairInterface[int, int]
	// IncrementUnHealthyFailCount 增加不健康失败计数
	IncrementUnHealthyFailCount()
	// ResetUnHealthyFailCount 重置不健康失败计数
	ResetUnHealthyFailCount()
	// GetUnHealthyFailCount 返回不健康失败计数
	GetUnHealthyFailCount() int64
	// GetUpStreamServerURL 返回上游服务器的URL
	GetUpStreamServerURL() string
	// ActiveHealthyCheck 执行主动健康检查，返回检查是否通过和可能的错误
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

	// SetHealthCheckIntervalMs 设置健康状态的缓存最大年龄。
	// 参数:
	//   int64 - 表示健康状态的缓存的最大年龄（单位：毫秒）。
	SetHealthyCheckInterval(int64)
	// 设置逻辑实现

	// GetHealthCheckIntervalMs 获取健康状态的缓存最大年龄。
	// 返回值:
	//   int64 - 健康状态的缓存的最大年龄（单位：毫秒）。
	GetHealthyCheckInterval() int64
	// 获取逻辑实现
	/*
		SetUnHealthyFailDurationMs 设置不健康状态失败持续时间。

		参数:
		- duration: int64类型，表示不健康状态失败的持续时间。

		返回值:
		- 无

		该函数用于设置一个指定的时间长度，在此时间长度内，如果组件或系统的状态持续不健康，则可能导致操作失败或系统视为异常。
	*/
	SetUnHealthyFailDurationMs(int64)
	// 设置逻辑实现

	// GetHealthCheckIntervalMs 获取健康状态的缓存最大年龄。
	// 返回值:
	//   int64 - 健康状态的缓存的最大年龄（单位：毫秒）。
	GetUnHealthyFailDurationMs() int64

	SetUnHealthyFailMaxCount(int64)

	GetUnHealthyFailMaxCount() int64

	OnUpstreamFailure()
}
type LoadBalanceService interface {
	GetUpStreams() generic.MapInterface[string, LoadBalanceAndUpStream]

	//选择一个可用的上游服务器
	// 参数：
	SelectAvailableServer() (LoadBalanceAndUpStream, error)
	HealthyCheckStart()
	HealthyCheckRunning() bool
	HealthyCheckStop()
	GetIdentifier() string
}
