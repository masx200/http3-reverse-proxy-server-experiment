package dns

import (
	// "context"
	// "time"
	// h12_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/h12"
	print_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/print"

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
func DohClient(msg *dns.Msg, dohServer string) (r *dns.Msg, err error) {
	body, err := msg.Pack()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//http request doh
	res, err := http.Post(dohServer, "application/dns-message", strings.NewReader(string(body)))
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
