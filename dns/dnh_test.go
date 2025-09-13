package dns

import (
	"errors"
	"log"
	"testing"

	"github.com/miekg/dns"
)

func TestDOH(t *testing.T) {
	dohServer := []string{"https://doh.opendns.com/dns-query", "https://doh.cleanbrowsing.org/doh/security-filter/"}
	// dohServer := []string{"9.9.9.9:853", "1.1.1.1:853", "8.8.4.4:853", "dot.sb:853"}
	domain := "production.hello-word-worker-cloudflare.masx200.workers.dev"

	var chan3 = make(chan error, len(dohServer))

	for _, server := range dohServer {
		go func(dohServer string) {
			var e error = nil
			defer func() {
				chan3 <- e
			}()
			// client := new(dns.Client)
			// client.Net = "tcp-tls"
			var tasks = []func() error{func() error {
				var msg = new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeAAAA)

				resp, err := DohClient(msg, dohServer)
				if err != nil {
					log.Println(err)

					return err
				}
				if resp.Rcode != dns.RcodeSuccess {
					log.Println(dns.RcodeToString[resp.Rcode])
					return errors.New("dns server " + dohServer + " response error not success")
				}
				if len(resp.Answer) == 0 {
					log.Println(dohServer + "-No AAAA records found")
					return errors.New(
						"dns server  response error No AAAA records found",
					)
				}
				log.Println(dohServer + "-" + resp.String())
				for _, answer := range resp.Answer {
					log.Println(answer)
					if a, ok := answer.(*dns.AAAA); ok {
						log.Printf(dohServer+"-Aaaa record for %s: %s\n", domain, a.AAAA)
					}
				}
				return nil
			}, func() error {
				msg := new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeA)

				resp, err := DohClient(msg, dohServer)
				if err != nil {
					log.Println(err)
					return err
				}
				if resp.Rcode != dns.RcodeSuccess {
					log.Println(dns.RcodeToString[resp.Rcode])
					return errors.New("dns server " + dohServer + " response error not success")
				}

				if len(resp.Answer) == 0 {
					log.Println(dohServer + "-No A records found")
					return errors.New(
						"dns server  response error No A records found",
					)
				}
				log.Println(dohServer + "-" + resp.String())
				for _, answer := range resp.Answer {
					log.Println(answer)
					if a, ok := answer.(*dns.A); ok {
						log.Printf(dohServer+"-A record for %s: %s\n", domain, a.A)
					}
				}
				return nil
			}, func() error {
				var msg = new(dns.Msg)
				msg.SetQuestion(domain+".", dns.TypeHTTPS)

				resp, err := DohClient(msg, dohServer)
				if err != nil {
					log.Println(err)
					return err
				}
				if resp.Rcode != dns.RcodeSuccess {
					log.Println(dns.RcodeToString[resp.Rcode])
					return errors.New("dns server " + dohServer + " response error not success")
				}
				if len(resp.Answer) == 0 {
					log.Println(dohServer + "-No HTTPS records found")
					return errors.New(
						"dns server  response HTTPS No AAAA records found",
					)
				}
				log.Println(dohServer + "-" + resp.String())
				for _, answer := range resp.Answer {
					log.Println(answer)
					if a, ok := answer.(*dns.HTTPS); ok {
						log.Printf(dohServer+"-https record for %s: \n", domain)

						for _, v := range a.SVCB.Value {

							log.Printf("%s", v.Key().String()+"="+v.String())
							log.Println()
						}
					}
				}
				return nil
			}}
			var chan2 = make(chan error, len(tasks))

			for _, task := range tasks {
				go func(task func() error) {
					var e error = nil
					defer func() {
						chan2 <- e
					}()
					e = task()
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
				e1 := <-chan2
				if e1 != nil {
					log.Fatal(e)
				}
				e = e1
				// <-chan2
				// <-chan2
			}

		}(server)

	}
	for range dohServer {
		e := <-chan3
		if e != nil {
			t.Fatal(e)
		}
	}
}
