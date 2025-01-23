package load_balance

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	// "strings"
	"sync"
	// "net/url"

	"github.com/masx200/http3-reverse-proxy-server-experiment/generic"
	h12_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/h12"
	print_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/print"
	optional "github.com/moznion/go-optional"
)

func PrintResponse(resp *http.Response) {
	print_experiment.PrintResponse(resp)
}

// ActiveHealthyCheckDefault 执行一个主动的健康检查，默认使用HEAD请求对给定的URL进行检查。
// 参数:
// - RoundTripper: 实现了http.RoundTripper接口的对象，用于发送HTTP请求。如果为空，将使用默认的http.Transport。
// - url: 需要进行健康检查的URL地址。
// 返回值:
// - bool: 检查结果，如果服务被认为是健康的则返回true，否则返回false。
// - error: 执行过程中遇到的任何错误。
func ActiveHealthyCheckDefault(RoundTripper http.RoundTripper, url string, method string, statusCodeMin int, statusCodeMax int) (bool, error) {

	// client := &http.Client{
	// 	CheckRedirect: func(req *http.Request, via []*http.Request) error {
	// 		return http.ErrUseLastResponse
	// 	},
	// }

	// // 使用提供的RoundTripper设置HTTP客户端的传输层。
	// client.Transport = RoundTripper
	// 创建一个新的HEAD请求。
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return false, err // 如果请求创建失败，则返回错误。
	}
	PrintRequest(req) // 打印请求信息，用于调试。

	resp, err := RoundTripper.RoundTrip(req) // 发送请求并获取响应。
	if err != nil {
		return false, err // 如果请求过程中有错误，则返回错误。
	}
	defer resp.Body.Close() // 确保在函数返回前关闭响应体。

	PrintResponse(resp) // 打印响应信息，用于调试。

	// 检查响应是否表明服务是健康的。
	return HealthyResponseCheckSuccess(resp, statusCodeMin, statusCodeMax)
	// 如果服务健康，则返回true，否则返回false。
}

func HealthyResponseCheckSuccess(response *http.Response, statusCodeMin int, statusCodeMax int) (bool, error) {
	// 检查响应状态码是否小于500
	if !(response.StatusCode >= statusCodeMin && response.StatusCode < statusCodeMax) {
		return false, errors.New("StatusCode" + fmt.Sprint(response.StatusCode) + "is not success")
	}
	return true, nil
}

func PrintRequest(req *http.Request) {
	print_experiment.PrintRequest(req)
}

// NewSingleHostHTTP12ClientOfAddress 创建一个指定地址的单主机HTTP客户端实例。
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
func NewSingleHostHTTP12ClientOfAddress(Identifier string, UpStreamServerURL string /*  ServerAddress string */, options ...func(*SingleHostHTTP12ClientOfAddress)) (LoadBalanceAndUpStream, error) {
	var ServerAddress, err = ExtractHostname(UpStreamServerURL)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// transport := h12_experiment.CreateHTTP12TransportWithIP(ServerAddress)
	// 初始化SingleHostHTTPClientOfAddress实例，并设置其属性值。

	m := &SingleHostHTTP12ClientOfAddress{
		Identifier:              Identifier,
		ActiveHealthyChecker:    ActiveHealthyCheckDefault,              // 使用默认的主动健康检查器
		PassiveUnHealthyChecker: HealthyResponseCheckDefault,            // 使用默认的健康响应检查器
		UpStreamServerURL:       UpStreamServerURL,                      // 设置上游服务器URL
		GetServerAddress:        func() string { return ServerAddress }, // 设置服务端地址
		IsHealthy:               true,                                   // 初始状态设为健康
		// RoundTripper:           transport,                   // 使用默认的传输器
		HealthCheckIntervalMs:   HealthCheckIntervalMsDefault,
		unHealthyFailDurationMs: unHealthyFailDurationMsDefault,
		UnHealthyFailMaxCount:   UnHealthyFailMaxCountDefault,
	}
	m.ServerConfigCommon = ServerConfigImplementConstructor(m.Identifier, m.UpStreamServerURL, m)
	/* 需要把transport保存起来,防止一个请求一个连接的情况速度会很慢 */

	// if strings.HasPrefix(m.UpStreamServerURL, "https") {
	/* 按照加密和不加密进行选择http2还是http1 */
	var h2rtcl = h12_experiment.CreateHTTP12TransportWithIPGetter(func() string {
		return m.GetServerAddress()
	})
	m.RoundTripper = h2rtcl
	m.Closer = func() error {
		return h2rtcl.Close()
	}
	// } else {
	// 	m.RoundTripper = h12_experiment.CreateHTTP1TransportWithIPGetter(func() string {
	// 		return m.GetServerAddress()
	// 	})
	// }

	for _, option := range options {
		option(m)
	}
	return m, nil
}

const unHealthyFailDurationMsDefault = 10 * 1000

// SingleHostHTTP12ClientOfAddress 是一个针对单个主机的HTTP客户端结构体，用于管理与特定地址的HTTP通信。
type SingleHostHTTP12ClientOfAddress struct {
	ServerConfigCommon      ServerConfigCommon
	unHealthyFailDurationMs int64
	HealthCheckIntervalMs   int64
	GetServerAddress        func() string                                                                                                       // 服务器地址，指定客户端要连接的HTTP服务器的地址。
	ActiveHealthyChecker    func(RoundTripper http.RoundTripper, url string, method string, statusCodeMin int, statusCodeMax int) (bool, error) // 活跃健康检查函数，用于检查给定的传输和URL是否健康。
	Identifier              string                                                                                                              // 标识符，用于标识此HTTP客户端的唯一字符串。
	HealthMutex             sync.Mutex
	IsHealthy               bool                                                                                        // 健康状态，标识当前客户端是否被视为健康。
	PassiveUnHealthyChecker func(response *http.Response, UnHealthyStatusMin int, UnHealthyStatusMax int) (bool, error) // 健康响应检查函数，用于基于HTTP响应检查客户端的健康状态。
	RoundTripper            http.RoundTripper                                                                           // HTTP传输，用于执行HTTP请求的实际传输。
	UpStreamServerURL       string                                                                                      // 上游服务器URL，指定客户端将请求转发到的上游服务器的地址。
	UnHealthyFailMaxCount   int64
	Closer                  func() error
}

// GetActiveHealthyCheckEnabled implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) GetActiveHealthyCheckEnabled() bool {
	return l.ServerConfigCommon.GetActiveHealthyCheckEnabled()
}

// GetPassiveHealthyCheckEnabled implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) GetPassiveHealthyCheckEnabled() bool {
	return l.ServerConfigCommon.GetPassiveHealthyCheckEnabled()
}

// SetActiveHealthyCheckEnabled implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) SetActiveHealthyCheckEnabled(e bool) {
	l.GetServerConfigCommon().SetActiveHealthyCheckEnabled(e)
}

// SetPassiveHealthyCheckEnabled implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) SetPassiveHealthyCheckEnabled(e bool) {
	l.GetServerConfigCommon().SetPassiveHealthyCheckEnabled(e)
}

// Close implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) Close() error {
	return l.Closer()
}

// GetServerConfigCommon implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) GetServerConfigCommon() ServerConfigCommon {
	return l.ServerConfigCommon
}

// GetUnHealthyFailMaxCount implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) GetUnHealthyFailMaxCount() int64 {
	return l.UnHealthyFailMaxCount
}

// SetUnHealthyFailMaxCount implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) SetUnHealthyFailMaxCount(count int64) {
	l.UnHealthyFailMaxCount = count
}

// GetLoadBalanceService implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) GetLoadBalanceService() optional.Option[LoadBalanceService] {
	return optional.None[LoadBalanceService]()
}

// GetHealthyCheckInterval implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) GetHealthyCheckInterval() int64 {
	return l.HealthCheckIntervalMs
}

// GetUnHealthyFailDurationMs implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) GetUnHealthyFailDurationMs() int64 {
	return l.unHealthyFailDurationMs
}

// SetHealthyCheckInterval implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) SetHealthyCheckInterval(interval int64) {
	l.HealthCheckIntervalMs = interval
}

// SetUnHealthyFailDurationMs implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) SetUnHealthyFailDurationMs(Duration int64) {
	l.unHealthyFailDurationMs = Duration
}

// GetHealthCheckIntervalMs implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) GetHealthCheckIntervalMs() int64 {

	return l.HealthCheckIntervalMs
}

// SetHealthCheckIntervalMs implements LoadBalanceAndUpStream.
func (l *SingleHostHTTP12ClientOfAddress) SetHealthCheckIntervalMs(maxAge int64) {

	l.HealthCheckIntervalMs = maxAge
}

// ActiveHealthyCheck 执行活跃的健康检查。
// 实现了 LoadBalanceAndUpStream 接口。
// 返回值：检查是否成功（bool类型）和可能发生的错误（error类型）。
func (l *SingleHostHTTP12ClientOfAddress) ActiveHealthyCheck() (bool, error) {
	return l.ActiveHealthyChecker(l, l.ServerConfigCommon.GetActiveHealthyCheckURL(), l.ServerConfigCommon.GetActiveHealthyCheckMethod(), l.ServerConfigCommon.GetActiveHealthyCheckStatusCodeRange().GetFirst(), l.ServerConfigCommon.GetActiveHealthyCheckStatusCodeRange().GetSecond())
}

// GetIdentifier 获取标识符。
// 实现了 LoadBalanceAndUpStream 接口。
// 返回值：此客户端的唯一标识符（string类型）。
func (l *SingleHostHTTP12ClientOfAddress) GetIdentifier() string {
	return l.Identifier
}

// GetHealthy 获取健康状态。
// 实现了 LoadBalanceAndUpStream 接口。
// 返回值：当前客户端是否处于健康状态（bool类型）。
func (l *SingleHostHTTP12ClientOfAddress) GetHealthy() bool {
	l.HealthMutex.Lock()
	defer l.HealthMutex.Unlock()
	return l.IsHealthy
}

// PassiveUnHealthyCheck 对HTTP响应进行健康状态检查。
// 实现了 LoadBalanceAndUpStream 接口。
// 参数：HTTP响应（*http.Response类型）。
// 返回值：检查结果是否健康（bool类型）和可能发生的错误（error类型）。
func (l *SingleHostHTTP12ClientOfAddress) PassiveUnHealthyCheck(response *http.Response) (bool, error) {
	return l.PassiveUnHealthyChecker(response, l.ServerConfigCommon.GetPassiveUnHealthyCheckStatusCodeRange().GetFirst(), l.ServerConfigCommon.GetPassiveUnHealthyCheckStatusCodeRange().GetSecond())
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
func HealthyResponseCheckDefault(response *http.Response, UnHealthyStatusMin int, UnHealthyStatusMax int) (bool, error) {

	// 检查响应状态码是否小于500
	if response.StatusCode >= UnHealthyStatusMin && response.StatusCode < UnHealthyStatusMax {
		return false, errors.New("StatusCode " + fmt.Sprint(response.StatusCode) + "   is greater than 500")
	}
	return true, nil
} // SingleHostHTTPClientOfAddress 是一个针对单个主机地址的HTTP客户端实现，
// 实现了LoadBalanceAndUpStream接口，用于负载均衡和上游服务管理。

// RoundTrip 实现了LoadBalanceAndUpStream接口的RoundTrip方法，
// 用于执行HTTP请求。
// 参数request为待发送的HTTP请求。
// 返回值为执行请求后的HTTP响应及可能发生的错误。
func (l *SingleHostHTTP12ClientOfAddress) RoundTrip(request *http.Request) (*http.Response, error) {
	var req = request
	var upurl, err = url.Parse(l.UpStreamServerURL)
	if err != nil {
		return nil, err
	}
	/* 要修改请求的URL和Header的Scheme和Host */
	req.URL.Scheme = upurl.Scheme
	req.URL.Host = upurl.Host
	req.Header.Set("Host", upurl.Host)
	req.Host = upurl.Host
	PrintRequest(req)
	// return l.RoundTripper.RoundTrip(request)
	/* 需要把transport保存起来,防止一个请求一个连接的情况速度会很慢 */
	return l.RoundTripper.RoundTrip(req)
}

// SelectAvailableServer 实现了LoadBalanceAndUpStream接口的SelectAvailableServer方法，
// 用于选择可用的服务实例。
// 返回值为可用的服务实例（此处始终为自身）及可能发生的错误。
func (l *SingleHostHTTP12ClientOfAddress) SelectAvailableServer() (LoadBalanceAndUpStream, error) {
	return nil, errors.New("no upstream available")
}

// SetHealthy 实现了LoadBalanceAndUpStream接口的SetHealthy方法，
// 用于设置客户端的健康状态。
// 参数healthy为true表示客户端健康，为false表示客户端不健康。
func (l *SingleHostHTTP12ClientOfAddress) SetHealthy(healthy bool) {
	l.HealthMutex.Lock()
	defer l.HealthMutex.Unlock()
	l.IsHealthy = healthy
}

// UpStreams 实现了LoadBalanceAndUpStream接口的UpStreams方法，
// 用于获取上游服务的集合。
// 此处因为是单主机客户端，所以返回空集合。
// 返回值为上游服务集合的可选类型，此处始终返回None。
func (l *SingleHostHTTP12ClientOfAddress) GetUpStreams() optional.Option[generic.MapInterface[string, LoadBalanceAndUpStream]] {
	return optional.None[generic.MapInterface[string, LoadBalanceAndUpStream]]()
}
