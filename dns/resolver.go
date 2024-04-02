package dns

import (
	"fmt"
	"sync"

	"github.com/miekg/dns"
)

type DnsResolverOptions struct {
	QueryCallback func(m *dns.Msg) (r *dns.Msg, err error)
	Domain        string

	HttpsPort int
}

func DnsResolver(queryCallback func(m *dns.Msg) (r *dns.Msg, err error), domain string, optionsCallBacks ...func(*DnsResolverOptions)) ([]string, error) {

	var options = &DnsResolverOptions{QueryCallback: queryCallback, Domain: domain, HttpsPort: 443}

	for _, optionsCallBack := range optionsCallBacks {
		optionsCallBack(options)
	}
	var wg sync.WaitGroup
	wg.Add(3)

	var results []string
	var resultsMutex sync.Mutex

	go func() {
		defer wg.Done()
		res, err := resolve(options, dns.TypeA)
		if err != nil {
			fmt.Printf("Error querying A record for %s: %v\n", options.Domain, err)
			return
		}
		resultsMutex.Lock()
		results = append(results, res...)
		resultsMutex.Unlock()
	}()

	go func() {
		defer wg.Done()
		res, err := resolve(options, dns.TypeAAAA)
		if err != nil {
			fmt.Printf("Error querying AAAA record for %s: %v\n", options.Domain, err)
			return
		}
		resultsMutex.Lock()
		results = append(results, res...)
		resultsMutex.Unlock()
	}()

	go func() {
		defer wg.Done()
		res, err := resolve(options, dns.TypeHTTPS)
		if err != nil {
			fmt.Printf("Error querying HTTPS record for %s: %v\n", options.Domain, err)
			return
		}
		resultsMutex.Lock()
		results = append(results, res...)
		resultsMutex.Unlock()
	}()

	wg.Wait()
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for %s", options.Domain)
	}
	return results, nil

}
func resolve(options *DnsResolverOptions, recordType uint16) ([]string, error) {
	m := &dns.Msg{}
	m.SetQuestion(dns.Fqdn(options.Domain), recordType)
	r, err := options.QueryCallback(m)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, answer := range r.Answer {
		switch record := answer.(type) {
		case *dns.A:
			results = append(results, fmt.Sprintf("A: %s", record.A))
		case *dns.AAAA:
			results = append(results, fmt.Sprintf("AAAA: %s", record.AAAA))
		case *dns.HTTPS:
			results = append(results, fmt.Sprintf("HTTPS: %s", record.Target))
		case *dns.CNAME:
			results = append(results, fmt.Sprintf("CNAME: %s", record.Target))
			res, err := DnsResolver(options.QueryCallback, record.Target)
			if err != nil {
				return nil, err
			}
			results = append(results, res...)
		}
	}
	return results, nil
}
