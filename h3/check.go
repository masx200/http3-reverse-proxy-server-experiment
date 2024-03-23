package h3

import (
	"fmt"
	"log"
	"strings"
	// "testing"

	dns_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/dns"
	"github.com/miekg/dns"
)

// CheckH3ViaDNS 通过DNS查询来检查指定域名和端口是否支持H3协议。
// domain: 需要检查的域名。
// port: 需要检查的端口。
// DOHServer: DNS-over-HTTPS服务器的地址。
// 返回值: 支持H3协议返回true，否则返回false。如果出现错误，将返回错误信息。
func CheckHttp3ViaDNS(domain string, port string, DOHServer string) (bool, error) {
	var records, err = DNSQueryHTTPS(domain, port, DOHServer)

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
	return false, fmt.Errorf("no H3 alpn records found")
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

// DNSQueryHTTPS 执行DNS查询以获取HTTPS服务记录。
//
// 参数:
// - domain: 需要查询的域名。
// - port: 目标端口，如果不为"443"，则会构建特定端口的查询域名。
// - DOHServer: DNS-over-HTTPS服务器地址。
//
// 返回值:
// - []dns.SVCB: 查询到的HTTPS服务记录列表。
// - error: 查询过程中发生的任何错误。
func DNSQueryHTTPS(domain string, port string, DOHServer string) ([]dns.SVCB, error) {
	var msg = new(dns.Msg)
	var service_domain = domain
	if port != "443" {
		service_domain = fmt.Sprintf("_%s._https.", port) + domain
	}
	msg.SetQuestion(service_domain+".", dns.TypeHTTPS)

	resp, err := DohClient(msg, DOHServer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if resp.Rcode != dns.RcodeSuccess {
		log.Printf("DNS query failed: %s ", dns.RcodeToString[resp.Rcode]+" "+DOHServer+"\n")
		return nil, fmt.Errorf(
			"DNS query failed: %s ", dns.RcodeToString[resp.Rcode]+" "+DOHServer)
	}
	if len(resp.Answer) == 0 {
		log.Println(DOHServer + "-No HTTPS records found")
		return nil, fmt.Errorf(
			"No HTTPS records found" + " " + DOHServer)
	}
	log.Println(DOHServer + "-" + resp.String())
	var result []dns.SVCB
	for _, answer := range resp.Answer {
		log.Println(answer)
		if a, ok := answer.(*dns.HTTPS); ok {
			fmt.Printf(DOHServer+"-https record for %s: \n", domain)
			result = append(result, a.SVCB)

		}
	}
	if len(result) == 0 {
		log.Println(DOHServer + "-No HTTPS records found")
		return nil, fmt.Errorf(
			"No HTTPS records found" + " " + DOHServer)
	}
	return result, nil

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
