package h3

import (
	"fmt"
	"log"

	dns_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/dns"
	"github.com/miekg/dns"
)

func Check(domain string, port string, DOHServer string) (bool, error) {}
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
	return result, nil

}

func DohClient(msg *dns.Msg, DOHServer string) (r *dns.Msg, err error) {
	return dns_experiment.DohClient(msg, DOHServer)
}
