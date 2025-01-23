package h3

import (
	// "context"
	// "crypto/tls"
	"errors"
	"fmt"
	"log"

	// "net"
	"net/http"
	"strings"

	// "testing"

	print_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/print"
	// "github.com/quic-go/quic-go"
	// "github.com/quic-go/quic-go/http3"
	altsvc "github.com/ebi-yade/altsvc-go"
	dns_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/dns"
	"github.com/miekg/dns"
)

// getAltSvc 发送一个 HEAD 请求到指定的 URL 并返回 Alt-Svc 响应头的内容。
// 如果成功获取到 Alt-Svc 头，函数将返回其值和 nil 作为错误；
// 如果遇到错误，将返回空字符串和相应的错误信息。
//
// 参数:
//
//	url: 需要发送 HEAD 请求的 URL 字符串。
//
// 返回值:
//
//	string: Alt-Svc 响应头的内容。如果未找到该头，则返回 "Not found"。
//	error: 发送请求或处理响应时遇到的错误，如果一切顺利则为 nil。
func getAltSvc(url string) (string, error) {
	resp, err := http.Head(url)
	if err != nil {
		return "", errors.New("failed to send HEAD request: " + err.Error())
	}

	PrintResponse(resp)
	defer resp.Body.Close()

	// 检查状态码是否成功（200 等）
	if resp.StatusCode >= 500 {
		return "", errors.New("received non-success status code: " + fmt.Sprint(resp.StatusCode))
	}

	// 获取 Alt-Svc 响应头
	altSvc := resp.Header.Get("Alt-Svc")
	if altSvc == "" {
		return "Not found", errors.New("Alt-Svc header not found")
	}
	log.Println("altSvc", altSvc)
	return altSvc, nil
}

func PrintResponse(resp *http.Response) {
	print_experiment.PrintResponse(resp)
}

// CheckHttp3ViaHttp2 通过HTTP/2检查特定域名和端口是否支持HTTP/3
// 参数:
// - domain: 需要检查的域名
// - port: 需要检查的端口号
// 返回值:
// - bool: 如果支持HTTP/3，则返回true，否则返回false
// - error: 如果检查过程中遇到错误，则返回错误信息
func CheckHttp3ViaHttp2(domain string, port string) (bool, error) {
	var altSvc, err = getAltSvc(fmt.Sprintf("https://%s:%s", domain, port))
	if err != nil {
		return false, err
	}
	svc, err := altsvc.Parse(altSvc)
	if err != nil {
		return false, err
	}
	for _, data := range svc {

		if data.ProtocolID == "h3" && data.AltAuthority.Port == port {
			return true, nil

		}

	}
	return false, errors.New("alt-svc Not found h3 and port")
}

// CheckH3ViaDNS 通过DNS查询来检查指定域名和端口是否支持H3协议。
// domain: 需要检查的域名。
// port: 需要检查的端口。
// DOHServer: DNS-over-HTTPS服务器的地址。
// 返回值: 支持H3协议返回true，否则返回false。如果出现错误，将返回错误信息。
func CheckHttp3ViaDNS(domain string, port string, DOHServer string) (bool, error) {
	var records, err = dns_experiment.DNSQueryHTTPS(domain, port, DOHServer)

	if err != nil {
		log.Println(err)
		return false, err
	}
	for _, record := range records {

		if record.Priority != 0 {
			for _, value := range record.Value {
				if value.Key().String() == "alpn" {
					var protocols = strings.Split(value.String(), ",")

					if ContainsGeneric(protocols, "h3") {
						return true, nil
					}
				}
			}
		}
	}
	return false, errors.New("no H3 alpn records found")
} // ContainsGeneric 函数用于判断一个切片中是否包含某个元素。
// 该函数支持泛型，可以适用于任意实现了可比较接口（comparable）的类型。
// 参数：
//
//	slice []T - 一个泛型切片，其中 T 必须实现 comparable 接口。
//	element T - 需要查找的元素，其类型与切片元素类型相同。
//
// 返回值：
//
//	bool - 如果切片中包含指定元素，则返回 true；否则返回 false。
func ContainsGeneric[T comparable](slice []T, element T) bool {
	for _, e := range slice {
		if e == element {
			return true
		}
	}

	return false
}

// DohClient 是一个通过DOH（DNS over HTTPS）协议与DNS服务器进行通信的函数。
// 它封装了dns_experiment包中的同名函数，简化了与DNS服务器交互的流程。
//
// 参数：
// msg         - 指向dns.Msg的指针，包含要发送的DNS查询信息。
// DOHServer   - 字符串类型，表示DOH服务器的URL。
//
// 返回值：
// r           - 指向dns.Msg的指针，包含从DNS服务器接收到的响应信息。
// err         - 错误类型，如果在与DNS服务器通信过程中发生错误，则返回非nil的错误值。
func DohClient(msg *dns.Msg, DOHServer string) (r *dns.Msg, err error) {
	return dns_experiment.DohClient(msg, DOHServer)
}

// FetchHttp3WithIP 使用IP地址通过HTTP/3协议获取网络资源。
//
// 参数:
// ip: 要使用的IP地址。
// url: 要请求的URL。
//
// 返回值:
// *http.Response: 请求成功的响应对象。
// error: 请求过程中发生的任何错误。
func FetchHttp3WithIP(ip, url string) (*http.Response, error) {
	// 创建UDP连接以用于QUIC协议

	// 创建HTTP/3客户端
	client := &http.Client{
		Transport: CreateHTTP3TransportWithIP(ip),
	}

	// 发起HTTP请求
	return client.Get(url)
}
