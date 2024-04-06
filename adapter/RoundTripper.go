package adapter

import "net/http"

// RoundTripTransport 是一个实现了 http.RoundTripper 接口的类型，
// 允许自定义HTTP请求的传输行为。
type RoundTripTransport func(*http.Request) (*http.Response, error) // roundTrip 是一个函数，它执行HTTP请求的传输，并返回响应和可能的错误。

// RoundTrip 是 http.RoundTripper 接口要求的方法，用于执行HTTP请求。
// 它简单地调用了结构体中的 roundTrip 函数，传递给它一个HTTP请求，并返回该请求的响应及可能的错误。
func (r RoundTripTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}

type HTTPRoundTripperAndCloserImplement struct {
	RoundTripper func(req *http.Request) (*http.Response, error)
	Closer       func() error
}

func (m *HTTPRoundTripperAndCloserImplement) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripper(req)
}
func (m *HTTPRoundTripperAndCloserImplement) Close() error {
	return m.Closer()
}

// func init() {
// 	var _ HTTPRoundTripperAndCloserInterface = &HTTPRoundTripperAndCloserImplement{}
// }

type HTTPRoundTripperAndCloserInterface interface {
	RoundTrip(req *http.Request) (*http.Response, error)
	Close() error
}
