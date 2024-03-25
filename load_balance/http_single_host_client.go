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
func NewSingleHostHTTPClientOfAddress(identifier string, UpStreamServerURL string, ServerAddress string) LoadBalanceAndUpStream {
	return &SingleHostHTTPClientOfAddress{
		Identifier:               identifier,
		ActiveHealthyChecker:     ActiveHealthyCheckDefault,
		IsHealthyResponseChecker: IsHealthyResponseDefault,
		UpStreamServerURL:        UpStreamServerURL,
		ServerAddress:            ServerAddress,
		isHealthy:                true,
		RoundTripper:             http.DefaultTransport,
	}
}

type SingleHostHTTPClientOfAddress struct {
	ServerAddress            string
	ActiveHealthyChecker     func(RoundTripper http.RoundTripper, url string) (bool, error)
	Identifier               string
	isHealthy                bool
	IsHealthyResponseChecker func(response *http.Response) (bool, error)
	RoundTripper             http.RoundTripper
	UpStreamServerURL        string
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
