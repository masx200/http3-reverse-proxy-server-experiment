package main

import (
	"fmt"
	"io"

	"log"
	"net/http"
	"strings"

	"github.com/miekg/dns"
)

func dohClient(msg *dns.Msg, dohServer string,
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
		log.Println("http status code is not 200")
		return nil, fmt.Errorf("http status code is not 200")
	}
	//利用ioutil包读取百度服务器返回的数据
	data, err := io.ReadAll(res.Body)
	res.Body.Close() //一定要记得关闭连接
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
func main() {
	dohServer := []string{"https://deno-dns-over-https-server-5ehq9rg3chgf.deno.dev/dns-query", "https://nextjs-reverse-proxy-middleware-masx2.netlify.app/token/4yF6nSCifSLs8lfkb4t8OWP69kfpgiun/https/doh-cache-worker-cf.masx200.workers.dev/dns-query", "https://nextjs-reverse-proxy-middleware.onrender.com/token/4yF6nSCifSLs8lfkb4t8OWP69kfpgiun/https/doh-cache-worker-cf.masx200.workers.dev/dns-query"}
	// dohServer := []string{"9.9.9.9:853", "1.1.1.1:853", "8.8.4.4:853", "dot.sb:853"}
	domain := "production.hello-word-worker-cloudflare.masx200.workers.dev"

	var chan3 = make(chan struct{}, len(dohServer))

	for _, v := range dohServer {
		go func(dohServer string) {
			defer func() {
				chan3 <- struct{}{}
			}()
			// client := new(dns.Client)
			// client.Net = "tcp-tls"

			var chan2 = make(chan struct{}, 3)
			go func() {
				defer func() {
					chan2 <- struct{}{}
				}()
				msg := new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeA)

				resp, err := dohClient(msg, dohServer)
				if err != nil {
					log.Println(err)
					return
				}
				if resp.Rcode != dns.RcodeSuccess {
					log.Println(dns.RcodeToString[resp.Rcode])
					return
				}

				if len(resp.Answer) == 0 {
					log.Println("No A records found")
					return
				}

				for _, answer := range resp.Answer {
					log.Println(answer)
					if a, ok := answer.(*dns.A); ok {
						fmt.Printf(dohServer+"-A record for %s: %s\n", domain, a.A)
					}
				}

			}()

			go func() {
				defer func() {
					chan2 <- struct{}{}
				}()
				var msg = new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeAAAA)

				resp, err := dohClient(msg, dohServer)
				if err != nil {
					log.Println(err)
					return
				}
				if resp.Rcode != dns.RcodeSuccess {
					log.Println(dns.RcodeToString[resp.Rcode])
					return
				}
				if len(resp.Answer) == 0 {
					log.Println("No AAAA records found")
					return
				}

				for _, answer := range resp.Answer {
					log.Println(answer)
					if a, ok := answer.(*dns.AAAA); ok {
						fmt.Printf(dohServer+"-Aaaa record for %s: %s\n", domain, a.AAAA)
					}
				}

			}()
			go func() {
				defer func() {
					chan2 <- struct{}{}
				}()
				var msg = new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeHTTPS)

				resp, err := dohClient(msg, dohServer)
				if err != nil {
					log.Println(err)
					return
				}
				if resp.Rcode != dns.RcodeSuccess {
					log.Println(dns.RcodeToString[resp.Rcode])
					return
				}
				if len(resp.Answer) == 0 {
					log.Println("No HTTPS records found")
					return
				}

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
			}()
			<-chan2
			<-chan2
			<-chan2

		}(v)

	}
	for range dohServer {
		<-chan3
	}
}
