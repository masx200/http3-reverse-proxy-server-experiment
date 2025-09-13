package h12

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/masx200/http3-reverse-proxy-server-experiment/adapter"
	"golang.org/x/net/http2"
	// "golang.org/x/net/http2"
)

func FetchHttp2WithIP(ip, url string) (*http.Response, error) {
	transport := CreateHTTP12TransportWithIP(ip)
	client := &http.Client{
		Transport: transport}
	return client.Get(url)
}

// CreateHTTP12TransportWithIP 创建一个http.Transport实例，该实例通过指定的IP地址进行网络连接。
// 这对于需要强制通过特定IP地址访问HTTP服务的情况非常有用。
//
// 参数:
//
//	ip string - 用于建立连接的IP地址。
//
// 返回值:
//
//	*http.Transport - 配置好的http.Transport指针，可用于http.Client或其他需要http.Transport的场合。
func CreateHTTP12TransportWithIP(ip string) http.RoundTripper {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second, // 设置拨号超时时间为30秒
		KeepAlive: 30 * time.Second, // 设置保持活动状态的间隔为30秒
	}

	// 返回配置好的http.Transport实例
	return &http.Transport{
		ForceAttemptHTTP2: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr) // 从地址中分解出主机和端口
			if err != nil {
				return nil, err
			}
			// 使用指定的IP地址拨号连接
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if err != nil {
				return nil, err
			}
			// 打印连接成功信息
			log.Println("连接成功http1", host, port, conn.LocalAddr(), conn.RemoteAddr())
			return conn, err
		},
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr) // TLS连接时同样分解地址
			if err != nil {
				return nil, err
			}
			// 拨号并配置TLS连接
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if err != nil {
				return nil, err
			}
			// 打印TLS连接成功信息
			log.Println("连接成功http2", host, port, conn.LocalAddr(), conn.RemoteAddr())
			// 返回配置好的TLS连接
			return tls.Client(conn, &tls.Config{
				ServerName: host, // 使用原始域名，而不是IP地址，这对于证书匹配很重要
				// 如果需要，可以在这里配置TLS的其他选项，比如跳过证书验证
			}), nil
		},
	}

}

// CreateHTTP12TransportWithIPGetter 创建一个自定义的http.Transport实例，该实例允许通过getter函数动态获取IP地址来进行连接，适用于需要手动指定连接IP的场景。
// getter: 一个函数，用于获取要使用的IP地址。该函数会在每次建立连接时被调用。
// 返回值: 配置好的http.RoundTripper接口，即http.Transport实例，可直接用于http.Client中。
func CreateHTTP12TransportWithIPGetter(getter func() string) adapter.HTTPRoundTripperAndCloserInterface {
	/* 需要把connection保存起来,防止一个请求一个连接的情况速度会很慢 */
	dialer := &net.Dialer{
		Timeout:   30 * time.Second, // 设置拨号超时时间为30秒
		KeepAlive: 30 * time.Second, // 设置保持活动状态的间隔为30秒
	}

	// 返回配置好的http.Transport实例
	var roundTripper = &http.Transport{
		ForceAttemptHTTP2: true,
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {

			host, port, err := net.SplitHostPort(addr) // TLS连接时同样分解地址
			if err != nil {
				return nil, err
			}
			var cfg *tls.Config = &tls.Config{NextProtos: []string{"h2", "http/1.1"}, ServerName: host}
			// 拨号并配置TLS连接
			var ip = getter()
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if err != nil {
				log.Println("连接失败http2", host, port)
				return nil, err
			}
			// 打印TLS连接成功信息
			log.Println("连接成功http2", host, port, conn.LocalAddr(), conn.RemoteAddr())
			// 返回配置好的TLS连接
			return tls.Client(conn, cfg), /* &tls.Config{
					ServerName: host, // 使用原始域名，而不是IP地址，这对于证书匹配很重要
					// 如果需要，可以在这里配置TLS的其他选项，比如跳过证书验证
				} */nil
		},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr) // 从地址中分解出主机和端口
			if err != nil {
				return nil, err
			}
			var ip = getter()
			// 使用指定的IP地址拨号连接
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if err != nil {
				log.Println("连接失败http1", host, port)
				return nil, err
			}
			// 打印连接成功信息
			log.Println("连接成功http1", host, port, conn.LocalAddr(), conn.RemoteAddr())
			return conn, err
		},
	}
	return &adapter.HTTPRoundTripperAndCloserImplement{RoundTripper: (func(r *http.Request) (*http.Response, error) {

		// 创建UDP连接，作为QUIC协议的基础。

		// 创建HTTP/3传输器，定制了Dial函数以使用指定的IP地址。

		// 使用定制的HTTP/3传输器进行HTTP请求的传输。
		return roundTripper.RoundTrip(r)
	}), Closer: func() error {
		// if transportquic != nil {
		// 	transportquic.Close()
		// }
		roundTripper.CloseIdleConnections()
		return nil
	}}
}

// CreateHTTP12TransportWithIPGetter 创建一个自定义的http.Transport实例，该实例允许通过getter函数动态获取IP地址来进行连接，适用于需要手动指定连接IP的场景。
// getter: 一个函数，用于获取要使用的IP地址。该函数会在每次建立连接时被调用。
// 返回值: 配置好的http.RoundTripper接口，即http.Transport实例，可直接用于http.Client中。
func CreateHTTP2TransportWithIPGetter(getter func() string) http.RoundTripper {
	/* 需要把connection保存起来,防止一个请求一个连接的情况速度会很慢 */
	dialer := &net.Dialer{
		Timeout:   30 * time.Second, // 设置拨号超时时间为30秒
		KeepAlive: 30 * time.Second, // 设置保持活动状态的间隔为30秒
	}

	// 返回配置好的http.Transport实例
	return &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr) // TLS连接时同样分解地址
			if err != nil {
				return nil, err
			}
			// 拨号并配置TLS连接
			var ip = getter()
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if err != nil {
				log.Println("连接失败http2", host, port)
				return nil, err
			}
			// 打印TLS连接成功信息
			log.Println("连接成功http2", host, port, conn.LocalAddr(), conn.RemoteAddr())
			// 返回配置好的TLS连接
			return tls.Client(conn, cfg), /* &tls.Config{
					ServerName: host, // 使用原始域名，而不是IP地址，这对于证书匹配很重要
					// 如果需要，可以在这里配置TLS的其他选项，比如跳过证书验证
				} */nil
		},
	}

}
