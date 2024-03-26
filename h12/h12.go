package h12

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

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
			fmt.Println("连接成功http1", host, port, conn.LocalAddr(), conn.RemoteAddr())
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
			fmt.Println("连接成功http2", host, port, conn.LocalAddr(), conn.RemoteAddr())
			// 返回配置好的TLS连接
			return tls.Client(conn, &tls.Config{
				ServerName: host, // 使用原始域名，而不是IP地址，这对于证书匹配很重要
				// 如果需要，可以在这里配置TLS的其他选项，比如跳过证书验证
			}), nil
		},
	}

}
