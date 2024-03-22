package main

import (
	"fmt"
	"log"

	"github.com/miekg/dns"
)

func main() {
	dnsServer := []string{"9.9.9.9:853", "1.1.1.1:853", "8.8.4.4:853", "dot.sb:853"}
	domain := "production.hello-word-worker-cloudflare.masx200.workers.dev"

	var chan3 = make(chan struct{}, len(dnsServer))

	for _, v := range dnsServer {
		go func(dnsServer string) {
			defer func() {
				chan3 <- struct{}{}
			}()
			client := new(dns.Client)
			client.Net = "tcp-tls"

			var chan2 = make(chan struct{}, 3)
			go func() {
				defer func() {
					chan2 <- struct{}{}
				}()
				msg := new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeA)

				resp, _, err := client.Exchange(msg, dnsServer)
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
						fmt.Printf(dnsServer+"-A record for %s: %s\n", domain, a.A)
					}
				}

			}()

			go func() {
				defer func() {
					chan2 <- struct{}{}
				}()
				var msg = new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeAAAA)

				var resp, _, err = client.Exchange(msg, dnsServer)
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
						fmt.Printf(dnsServer+"-Aaaa record for %s: %s\n", domain, a.AAAA)
					}
				}

			}()
			go func() {
				defer func() {
					chan2 <- struct{}{}
				}()
				var msg = new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeHTTPS)

				var resp, _, err = client.Exchange(msg, dnsServer)
				if err != nil {
					log.Fatal(err)
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
						fmt.Printf(dnsServer+"-https record for %s: \n", domain)

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
	for range dnsServer {
		<-chan3
	}
}
