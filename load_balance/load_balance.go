package load_balance

import (
	"net/http"

	"github.com/moznion/go-optional"
)

// LoadBalance 是一个负载均衡接口，它定义了如何对HTTP请求进行负载均衡转发。
// 其中包含了一个Map，用于映射域名到对应的UpStream。
type LoadBalanceAndUpStream interface {
	GetActiveHealthyCheckEnabled() bool
	SetActiveHealthyCheckEnabled(bool)
	GetPassiveHealthyCheckEnabled() bool
	SetPassiveHealthyCheckEnabled(bool)
	Close() error
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
