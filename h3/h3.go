package h3

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/masx200/http3-reverse-proxy-server-experiment/adapter"
	"github.com/miekg/dns"
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

// CreateHTTP3TransportWithIPGetter 创建一个带有自定义IP获取器的HTTP/3传输器。
// 此函数允许在每次HTTP请求时动态指定IP地址，用于建立QUIC连接。
//
// 参数:
// getter func() string - 一个函数，返回一个字符串形式的IP地址。
//
// 返回值:
// http.RoundTripper - 符合HTTP运输接口的定制HTTP/3传输器。
func CreateHTTP3TransportWithIPGetter(getter func() string) http.RoundTripper {
	var transport *quic.Transport
	var mutex sync.Mutex

	var mapconnection map[string]quic.EarlyConnection
	/* 需要把connection保存起来,防止一个请求一个连接的情况速度会很慢 */
	return adapter.RoundTripTransport(func(r *http.Request) (*http.Response, error) {

		// 创建UDP连接，作为QUIC协议的基础。

		// 创建HTTP/3传输器，定制了Dial函数以使用指定的IP地址。
		var transport = &http3.RoundTripper{
			Dial: func(ctx context.Context, addr string, tlsConf *tls.Config, quicConf *quic.Config) (quic.EarlyConnection, error) {

				mutex.Lock()
				defer mutex.Unlock()

				if mapconnection == nil {
					/* 需要初始化map */
					mapconnection = map[string]quic.EarlyConnection{}
				}
				var tr *quic.Transport
				if transport == nil {
					udpConn, err := net.ListenUDP("udp", nil)
					if err != nil {
						return nil, err
					}
					tr = &quic.Transport{Conn: udpConn}
					transport = tr
				} else {
					tr = transport
				}
				var ServerName = tlsConf.ServerName

				x := mapconnection[ServerName]
				if x != nil {

					fmt.Println("使用quic缓存连接", ServerName, addr, x.LocalAddr(), x.RemoteAddr())
					return x, nil
				}
				// 分解地址并替换为指定的IP地址。
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				var ip = getter()
				addr2 := net.JoinHostPort(ip, port)
				a, err := net.ResolveUDPAddr("udp", addr2)
				if err != nil {
					return nil, err
				}

				// 使用替换后的地址尝试建立QUIC连接。
				conn, err := tr.DialEarly(ctx, a, tlsConf, quicConf)
				if err != nil {
					fmt.Println("http3连接失败", ServerName, host, port /*  conn.LocalAddr(), conn.RemoteAddr() */)
					return nil, err
				}
				fmt.Println("http3连接成功", ServerName, host, port, conn.LocalAddr(), conn.RemoteAddr())
				mapconnection[ServerName] = conn
				return conn, err
			},
		}

		// 使用定制的HTTP/3传输器进行HTTP请求的传输。
		return transport.RoundTrip(r)
	})

}

// DohClient 是一个通过DOH（DNS over HTTPs）协议与DNS服务器进行通信的函数。
//
// 参数:
// msg: 代表DNS查询消息的dns.Msg对象。
// dohServer: 代表DOH服务器的URL字符串。
//
// 返回值:
// r: 代表DNS应答消息的dns.Msg对象。
// err: 如果过程中发生错误，则返回错误信息。
func DoHTTP3Client(msg *dns.Msg, dohttp3ServerURL string) (r *dns.Msg, err error) {
	client := &http.Client{
		Transport: &http3.RoundTripper{},
	}
	body, err := msg.Pack()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//http request doh
	res, err := client.Post(dohttp3ServerURL, "application/dns-message", strings.NewReader(string(body)))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//res.status check
	if res.StatusCode != 200 {
		log.Println("http status code is not 200 " + fmt.Sprintf("status code is %d", res.StatusCode))
		return nil, fmt.Errorf("http status code is not 200" + fmt.Sprintf("status code is %d", res.StatusCode))
	}

	//check content-type
	if res.Header.Get("Content-Type") != "application/dns-message" {
		log.Println("content-type is not application/dns-message " + res.Header.Get("Content-Type"))
		return nil, fmt.Errorf("content-type is not application/dns-message " + res.Header.Get("Content-Type"))
	}
	//利用ioutil包读取百度服务器返回的数据
	data, err := io.ReadAll(res.Body)
	defer res.Body.Close() //一定要记得关闭连接
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// log.Printf("%s", data)
	resp := &dns.Msg{}
	err = resp.Unpack(data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return resp, nil
}
