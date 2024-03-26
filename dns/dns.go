package dns

import (
	// "context"
	// "time"
	h12_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/h12"
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
func DohClient(msg *dns.Msg, dohServer string,
) (r *dns.Msg, err error) {
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
func Main() {
	dohServer := []string{"https://doh.pub/dns-query", "https://doh.360.cn/dns-query", "https://dns.alidns.com/dns-query"}
	// dohServer := []string{"9.9.9.9:853", "1.1.1.1:853", "8.8.4.4:853", "dot.sb:853"}
	domain := "production.hello-word-worker-cloudflare.masx200.workers.dev"

	var chan3 = make(chan struct{}, len(dohServer))

	for _, server := range dohServer {
		go func(dohServer string) {
			defer func() {
				chan3 <- struct{}{}
			}()
			// client := new(dns.Client)
			// client.Net = "tcp-tls"
			var tasks = []func(){func() {
				var msg = new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeAAAA)

				resp, err := DohClient(msg, dohServer)
				if err != nil {
					log.Println(err)
					return
				}
				if resp.Rcode != dns.RcodeSuccess {
					log.Println(dns.RcodeToString[resp.Rcode])
					return
				}
				if len(resp.Answer) == 0 {
					log.Println(dohServer + "-No AAAA records found")
					return
				}
				log.Println(dohServer + "-" + resp.String())
				for _, answer := range resp.Answer {
					log.Println(answer)
					if a, ok := answer.(*dns.AAAA); ok {
						fmt.Printf(dohServer+"-Aaaa record for %s: %s\n", domain, a.AAAA)
					}
				}
			}, func() {
				msg := new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeA)

				resp, err := DohClient(msg, dohServer)
				if err != nil {
					log.Println(err)
					return
				}
				if resp.Rcode != dns.RcodeSuccess {
					log.Println(dns.RcodeToString[resp.Rcode])
					return
				}

				if len(resp.Answer) == 0 {
					log.Println(dohServer + "-No A records found")
					return
				}
				log.Println(dohServer + "-" + resp.String())
				for _, answer := range resp.Answer {
					log.Println(answer)
					if a, ok := answer.(*dns.A); ok {
						fmt.Printf(dohServer+"-A record for %s: %s\n", domain, a.A)
					}
				}

			}, func() {
				var msg = new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeHTTPS)

				resp, err := DohClient(msg, dohServer)
				if err != nil {
					log.Println(err)
					return
				}
				if resp.Rcode != dns.RcodeSuccess {
					log.Println(dns.RcodeToString[resp.Rcode])
					return
				}
				if len(resp.Answer) == 0 {
					log.Println(dohServer + "-No HTTPS records found")
					return
				}
				log.Println(dohServer + "-" + resp.String())
				for _, answer := range resp.Answer {
					log.Println(answer)
					if a, ok := answer.(*dns.HTTPS); ok {
						fmt.Printf(dohServer+"-https record for %s: \n", domain)

						for _, v := range a.SVCB.Value {

							fmt.Printf("%s", v.Key().String()+"="+v.String())
							fmt.Println()
						}
					}
				}
			}}
			var chan2 = make(chan struct{}, len(tasks))

			for _, task := range tasks {
				go func(task func()) {
					defer func() {
						chan2 <- struct{}{}
					}()
					task()
				}(task)
			}
			// go func() {
			// 	defer func() {
			// 		chan2 <- struct{}{}
			// 	}()

			// }()

			// go func() {
			// 	defer func() {
			// 		chan2 <- struct{}{}
			// 	}()

			// }()
			// go func() {
			// 	defer func() {
			// 		chan2 <- struct{}{}
			// 	}()

			// }()

			for range tasks {
				<-chan2
				// <-chan2
				// <-chan2
			}

		}(server)

	}
	for range dohServer {
		<-chan3
	}
}

func FetchHttp2WithIP(ip, url string) (*http.Response, error) {
	transport := h12_experiment.CreateHTTP12TransportWithIP(ip)
	client := &http.Client{
		Transport: transport}
	return client.Get(url)
}
func PrintResponse(resp *http.Response) {
	print_experiment.PrintResponse(resp)
}
