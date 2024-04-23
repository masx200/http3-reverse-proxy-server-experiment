package load_balance

import (
	"net/http"

	"github.com/masx200/http3-reverse-proxy-server-experiment/generic"
)

type ServerConfigCommon interface {
	GetActiveHealthyCheckEnabled() bool
	SetActiveHealthyCheckEnabled(bool)
	GetPassiveHealthyCheckEnabled() bool
	SetPassiveHealthyCheckEnabled(bool)

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
	OnUpstreamHealthy()
}
