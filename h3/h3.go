package h3

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/masx200/http3-reverse-proxy-server-experiment/adapter"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// CreateHTTP3TransportWithIP 创建一个使用指定IP地址的HTTP/3传输。
//
// 参数:
//
//	ip string - 要使用的IP地址。
//
// 返回值:
//
//	http.RoundTripper - 一个实现了HTTP运输接口的对象，可以用于HTTP客户端进行请求。
func CreateHTTP3TransportWithIP(ip string) http.RoundTripper {
	return adapter.RoundTripTransport(func(r *http.Request) (*http.Response, error) {
		// 创建UDP连接，作为QUIC协议的基础。
		udpConn, err := net.ListenUDP("udp", nil)
		if err != nil {
			return nil, err
		}
		tr := quic.Transport{Conn: udpConn}

		// 创建HTTP/3传输器，定制了Dial函数以使用指定的IP地址。
		var transport = &http3.RoundTripper{
			Dial: func(ctx context.Context, addr string, tlsConf *tls.Config, quicConf *quic.Config) (quic.EarlyConnection, error) {
				// 分解地址并替换为指定的IP地址。
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				addr2 := net.JoinHostPort(ip, port)
				a, err := net.ResolveUDPAddr("udp", addr2)
				if err != nil {
					return nil, err
				}

				// 使用替换后的地址尝试建立QUIC连接。
				conn, err := tr.DialEarly(ctx, a, tlsConf, quicConf)
				if err != nil {
					fmt.Println("http3连接失败", host, port, conn.LocalAddr(), conn.RemoteAddr())
					return nil, err
				}
				fmt.Println("http3连接成功", host, port, conn.LocalAddr(), conn.RemoteAddr())
				return conn, err
			},
		}

		// 使用定制的HTTP/3传输器进行HTTP请求的传输。
		return transport.RoundTrip(r)
	})

}
