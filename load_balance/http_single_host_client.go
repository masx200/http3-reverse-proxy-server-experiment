package load_balance

import (
	"fmt"
	"net/http"
	// "net/url"

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
func ActiveHealthyCheckDefault(RoundTripper http.RoundTripper, url string) (bool, error) {

	client := &http.Client{}

	// 使用提供的RoundTripper设置HTTP客户端的传输层。
	client.Transport = RoundTripper
	// 创建一个新的HEAD请求。
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, err // 如果请求创建失败，则返回错误。
	}
	PrintRequest(req) // 打印请求信息，用于调试。

	resp, err := client.Do(req) // 发送请求并获取响应。
	if err != nil {
		return false, err // 如果请求过程中有错误，则返回错误。
	}
	defer resp.Body.Close() // 确保在函数返回前关闭响应体。

	PrintResponse(resp) // 打印响应信息，用于调试。

	// 检查响应是否表明服务是健康的。
	return IsHealthyResponseDefault(resp)
	// 如果服务健康，则返回true，否则返回false。
}

func PrintRequest(req *http.Request) {
	print_experiment.PrintRequest(req)
}

// NewSingleHostHTTPClientOfAddress 创建一个指定地址的单主机HTTP客户端实例。
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
func NewSingleHostHTTPClientOfAddress(identifier string, UpStreamServerURL string, ServerAddress string) LoadBalanceAndUpStream {
	// 初始化SingleHostHTTPClientOfAddress实例，并设置其属性值。
	return &SingleHostHTTPClientOfAddress{
		Identifier:               identifier,
		ActiveHealthyChecker:     ActiveHealthyCheckDefault, // 使用默认的主动健康检查器
		IsHealthyResponseChecker: IsHealthyResponseDefault,  // 使用默认的健康响应检查器
		UpStreamServerURL:        UpStreamServerURL,         // 设置上游服务器URL
		ServerAddress:            ServerAddress,             // 设置服务端地址
		isHealthy:                true,                      // 初始状态设为健康
		RoundTripper:             http.DefaultTransport,     // 使用默认的传输器
	}
}

// SingleHostHTTPClientOfAddress 是一个针对单个主机的HTTP客户端结构体，用于管理与特定地址的HTTP通信。
type SingleHostHTTPClientOfAddress struct {
	ServerAddress            string                                                         // 服务器地址，指定客户端要连接的HTTP服务器的地址。
	ActiveHealthyChecker     func(RoundTripper http.RoundTripper, url string) (bool, error) // 活跃健康检查函数，用于检查给定的传输和URL是否健康。
	Identifier               string                                                         // 标识符，用于标识此HTTP客户端的唯一字符串。
	isHealthy                bool                                                           // 健康状态，标识当前客户端是否被视为健康。
	IsHealthyResponseChecker func(response *http.Response) (bool, error)                    // 健康响应检查函数，用于基于HTTP响应检查客户端的健康状态。
	RoundTripper             http.RoundTripper                                              // HTTP传输，用于执行HTTP请求的实际传输。
	UpStreamServerURL        string                                                         // 上游服务器URL，指定客户端将请求转发到的上游服务器的地址。
}

// ActiveHealthyCheck implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) ActiveHealthyCheck() (bool, error) {
	return l.ActiveHealthyChecker(l, l.UpStreamServerURL)
}

// Identifier implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) GetIdentifier() string {
	return l.Identifier
}

// IsHealthy implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) IsHealthy() bool {
	return l.isHealthy
}

// IsHealthyResponse implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) IsHealthyResponse(response *http.Response) (bool, error) {
	return l.IsHealthyResponseChecker(response)
}
func IsHealthyResponseDefault(response *http.Response) (bool, error) {

	//check StatusCode<500
	if response.StatusCode >= 500 {
		return false, fmt.Errorf("StatusCode %d   is greater than 500", response.StatusCode)
	}
	return true, nil
}

// RoundTrip implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) RoundTrip(request *http.Request) (*http.Response, error) {
	return l.RoundTripper.RoundTrip(request)
}

// SelectAvailableServer implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) SelectAvailableServer() (LoadBalanceAndUpStream, error) {
	return l, nil
}

// SetHealthy implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) SetHealthy(healthy bool) {
	l.isHealthy = healthy
}

// UpStreams implements LoadBalanceAndUpStream.
func (l *SingleHostHTTPClientOfAddress) UpStreams() optional.Option[MapInterface[string, LoadBalanceAndUpStream]] {
	return optional.None[MapInterface[string, LoadBalanceAndUpStream]]()
}
