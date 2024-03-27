package dns

import (
	// "context"
	// "time"
	// h12_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/h12"
	"context"

	print_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/print"
	doq "github.com/tantalor93/doq-go/doq"

	// "crypto/tls"
	// "fmt"
	// "io"
	// "net"
	// "net/http"
	// "crypto/tls"
	"fmt"
	"io"

	"log"
	"net/http"
	"strings"

	"github.com/miekg/dns"
)

// DohClient 是一个通过DOH（DNS over HTTPs）协议与DNS服务器进行通信的函数。
//
// 参数:
// msg: 代表DNS查询消息的dns.Msg对象。
// dohServer: 代表DOH服务器的URL字符串。
//
// 返回值:
// r: 代表DNS应答消息的dns.Msg对象。
// err: 如果过程中发生错误，则返回错误信息。
func DohClient(msg *dns.Msg, dohServerURL string) (r *dns.Msg, err error) {
	body, err := msg.Pack()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//http request doh
	res, err := http.Post(dohServerURL, "application/dns-message", strings.NewReader(string(body)))
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

func PrintResponse(resp *http.Response) {
	print_experiment.PrintResponse(resp)
}

// DOQClient 是一个通过DOQ（DNS over QUIC）协议与DNS服务器进行通信的函数。
//
// 参数:
// msg 是一个包含DNS查询信息的dns.Msg结构体指针。
// dohServerURL 是一个字符串，表示DOQ服务器的URL。
//
// 返回值:
// 返回一个包含DNS应答信息的dns.Msg结构体指针和一个错误信息。
// 如果成功，错误信息为nil；如果发生错误，则返回相应的错误信息。
func DOQClient(msg *dns.Msg, dohServerURL string) (qA *dns.Msg, err error) {
	fmt.Println("dohServerURL", dohServerURL)
	// 从DOH服务器URL中提取服务器名称和端口信息。
	serverName, port, err := ExtractDOQServerDetails(dohServerURL)
	if err != nil {
		log.Println(err) // 记录提取详情时的错误
		return nil, err  // 如果有错误，返回nil和错误信息
	}
	var addr = fmt.Sprintf("%s:%s", serverName, port) // 格式化服务器地址
	fmt.Println("addr", addr)
	// 创建一个DOQ客户端
	client := doq.NewClient(addr, doq.Options{})
	// 发送DNS查询并获取应答
	respA, err := client.Send(context.Background(), msg)
	return respA, err // 返回DNS应答和可能的错误信息
}

// ExtractDOQServerDetails takes a DOQ server URL and returns the server name and port as separate strings.
func ExtractDOQServerDetails(doqServer string) (string, string, error) {
	parts := strings.Split(doqServer, "://")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid DOQ server format")
	}

	serverWithPort := parts[1]
	serverParts := strings.Split(serverWithPort, ":")
	if len(serverParts) != 2 {
		return "", "", fmt.Errorf("invalid server details, missing port")
	}

	serverName := serverParts[0]
	port := serverParts[1]
	return serverName, port, nil
}
