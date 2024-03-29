package load_balance

import (
	// "fmt"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/masx200/http3-reverse-proxy-server-experiment/generic"

	// "net/url"
	h3_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/h3"

	optional "github.com/moznion/go-optional"
)

const UnHealthyFailMaxCountDefault = 5
const HealthCheckIntervalDefault = 10 * 1000

// ActiveHealthyCheckDefault 执行一个主动的健康检查，默认使用HEAD请求对给定的URL进行检查。
// 参数:
// - RoundTripper: 实现了http.RoundTripper接口的对象，用于发送HTTP请求。如果为空，将使用默认的http.Transport。
// - url: 需要进行健康检查的URL地址。
// 返回值:
// - bool: 检查结果，如果服务被认为是健康的则返回true，否则返回false。
// - error: 执行过程中遇到的任何错误。

// NewSingleHostHTTP3ClientOfAddress 创建一个指定地址的单主机HTTP客户端实例。
//
// 参数:
//
//	identifier string - 客户端的标识符，用于区分不同的客户端实例。
//	UpStreamServerURL string - 上游服务器的URL，客户端将向该URL发送请求。
//	ServerAddress string - 服务端地址，指定客户端连接的服务端主机地址。
//
// 返回值:
//
//	LoadBalanceAndUpStream - 实现了负载均衡和上游服务选择的接口。
func NewSingleHostHTTP3ClientOfAddress(Identifier string, UpStreamServerURL string /*  ServerAddress string, */, options ...func(*SingleHostHTTP3ClientOfAddress)) (LoadBalanceAndUpStream, error) {
	var ServerAddress, err = ExtractHostname(UpStreamServerURL)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// transport := h3_experiment.CreateHTTP3TransportWithIP(ServerAddress)
	// // 初始化SingleHostHTTPClientOfAddress实例，并设置其属性值。

	m := &SingleHostHTTP3ClientOfAddress{
		HealthCheckInterval:     HealthCheckIntervalDefault, // 使用默认的健康缓存时间
		Identifier:              Identifier,
		ActiveHealthyChecker:    ActiveHealthyCheckDefault,              // 使用默认的主动健康检查器
		PassiveUnHealthyChecker: HealthyResponseCheckDefault,            // 使用默认的健康响应检查器
		UpStreamServerURL:       UpStreamServerURL,                      // 设置上游服务器URL
		GetServerAddress:        func() string { return ServerAddress }, // 设置服务端地址
		IsHealthy:               true,                                   // 初始状态设为健康
		// RoundTripper:           transport,                   // 使用默认的传输器
		UnHealthyFailDuration: UnHealthyFailDurationDefault,
		UnHealthyFailMaxCount: UnHealthyFailMaxCountDefault,
	}
	for _, option := range options {
		option(m)
	}
	return m, nil
}

// SingleHostHTTPClientOfAddress 是一个针对单个主机的HTTP客户端结构体，用于管理与特定地址的HTTP通信。
type SingleHostHTTP3ClientOfAddress struct {
	UnHealthyFailMaxCount   int64
	healthMutex             sync.Mutex
	UnHealthyFailDuration   int64
	HealthCheckInterval     int64
	GetServerAddress        func() string                                                  // 服务器地址，指定客户端要连接的HTTP服务器的地址。
	ActiveHealthyChecker    func(RoundTripper http.RoundTripper, url string) (bool, error) // 活跃健康检查函数，用于检查给定的传输和URL是否健康。
	Identifier              string                                                         // 标识符，用于标识此HTTP客户端的唯一字符串。
	IsHealthy               bool                                                           // 健康状态，标识当前客户端是否被视为健康。
	PassiveUnHealthyChecker func(response *http.Response) (bool, error)                    // 健康响应检查函数，用于基于HTTP响应检查客户端的健康状态。
	// RoundTripper           http.RoundTripper                                              // HTTP传输，用于执行HTTP请求的实际传输。
	UpStreamServerURL string // 上游服务器URL，指定客户端将请求转发到的上游服务器的地址。
}

// GetUnHealthyFailMaxCount implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP3ClientOfAddress) GetUnHealthyFailMaxCount() int64 {
	return l.UnHealthyFailMaxCount
}

// SetUnHealthyFailMaxCount implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP3ClientOfAddress) SetUnHealthyFailMaxCount(count int64) {
	l.UnHealthyFailMaxCount = count
}

// GetLoadBalanceService implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP3ClientOfAddress) GetLoadBalanceService() optional.Option[LoadBalanceService] {
	return optional.None[LoadBalanceService]()
}

// GetHealthyCheckInterval implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP3ClientOfAddress) GetHealthyCheckInterval() int64 {
	return l.HealthCheckInterval
}

// GetUnHealthyFailDuration implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP3ClientOfAddress) GetUnHealthyFailDuration() int64 {
	return l.UnHealthyFailDuration
}

// SetHealthyCheckInterval implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP3ClientOfAddress) SetHealthyCheckInterval(interval int64) {
	l.HealthCheckInterval = interval
}

// SetUnHealthyFailDuration implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP3ClientOfAddress) SetUnHealthyFailDuration(Duration int64) {
	l.UnHealthyFailDuration = Duration
}

// GetHealthCheckInterval implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP3ClientOfAddress) GetHealthCheckInterval() int64 {
	return l.HealthCheckInterval
}

// SetHealthCheckInterval implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP3ClientOfAddress) SetHealthCheckInterval(maxAge int64) {
	l.HealthCheckInterval = maxAge
}

// ActiveHealthyCheck 执行活跃的健康检查。
// 实现了 LoadBalanceAndUpStream 接口。
// 返回值：检查是否成功（bool类型）和可能发生的错误（error类型）。
func (l *SingleHostHTTP3ClientOfAddress) ActiveHealthyCheck() (bool, error) {
	return l.ActiveHealthyChecker(l, l.UpStreamServerURL)
}

// GetIdentifier 获取标识符。
// 实现了 LoadBalanceAndUpStream 接口。
// 返回值：此客户端的唯一标识符（string类型）。
func (l *SingleHostHTTP3ClientOfAddress) GetIdentifier() string {
	return l.Identifier
}

// GetHealthy 获取健康状态。
// 实现了 LoadBalanceAndUpStream 接口。
// 返回值：当前客户端是否处于健康状态（bool类型）。
func (l *SingleHostHTTP3ClientOfAddress) GetHealthy() bool {
	l.healthMutex.Lock()
	defer l.healthMutex.Unlock()
	return l.IsHealthy
}

// PassiveUnHealthyCheck 对HTTP响应进行健康状态检查。
// 实现了 LoadBalanceAndUpStream 接口。
// 参数：HTTP响应（*http.Response类型）。
// 返回值：检查结果是否健康（bool类型）和可能发生的错误（error类型）。
func (l *SingleHostHTTP3ClientOfAddress) PassiveUnHealthyCheck(response *http.Response) (bool, error) {
	return l.PassiveUnHealthyChecker(response)
}

// HealthyResponseCheckDefault 检查HTTP响应是否表示服务健康。
//
// 参数:
//
//	response *http.Response - 用于检查的HTTP响应对象。
//
// 返回值:
//
//	bool - 如果响应表示服务健康，则为true；否则为false。
//	error - 如果检查过程中遇到错误，则返回错误信息；否则为nil。
// SingleHostHTTPClientOfAddress 是一个针对单个主机地址的HTTP客户端实现，
// 实现了LoadBalanceAndUpStream接口，用于负载均衡和上游服务管理。

// RoundTrip 实现了LoadBalanceAndUpStream接口的RoundTrip方法，
// 用于执行HTTP请求。
// 参数request为待发送的HTTP请求。
// 返回值为执行请求后的HTTP响应及可能发生的错误。
func (l *SingleHostHTTP3ClientOfAddress) RoundTrip(request *http.Request) (*http.Response, error) {
	return h3_experiment.CreateHTTP3TransportWithIPGetter(func() string {
		return l.GetServerAddress()
	}).RoundTrip(request)
}

// SelectAvailableServer 实现了LoadBalanceAndUpStream接口的SelectAvailableServer方法，
// 用于选择可用的服务实例。
// 返回值为可用的服务实例（此处始终为自身）及可能发生的错误。
func (l *SingleHostHTTP3ClientOfAddress) SelectAvailableServer() (LoadBalanceAndUpStream, error) {
	return nil, fmt.Errorf("no upstream available")
}

// SetHealthy 实现了LoadBalanceAndUpStream接口的SetHealthy方法，
// 用于设置客户端的健康状态。
// 参数healthy为true表示客户端健康，为false表示客户端不健康。
func (l *SingleHostHTTP3ClientOfAddress) SetHealthy(healthy bool) {
	l.healthMutex.Lock()
	defer l.healthMutex.Unlock()
	l.IsHealthy = healthy
}

// UpStreams 实现了LoadBalanceAndUpStream接口的UpStreams方法，
// 用于获取上游服务的集合。
// 此处因为是单主机客户端，所以返回空集合。
// 返回值为上游服务集合的可选类型，此处始终返回None。
func (l *SingleHostHTTP3ClientOfAddress) GetUpStreams() optional.Option[generic.MapInterface[string, LoadBalanceAndUpStream]] {
	return optional.None[generic.MapInterface[string, LoadBalanceAndUpStream]]()
}
