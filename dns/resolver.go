package dns

import "github.com/miekg/dns"

type DnsResolverOptions struct {
	QueryCallback func(m *dns.Msg) (r *dns.Msg, err error)
	Domain        string

	HttpsPort int
}

func DnsResolver(queryCallback func(m *dns.Msg) (r *dns.Msg, err error), domain string, optionsCallBacks ...func(*DnsResolverOptions)) {

	var options = &DnsResolverOptions{QueryCallback: queryCallback, Domain: domain, HttpsPort: 443}

	for _, optionsCallBack := range optionsCallBacks {
		optionsCallBack(options)
	}
}
