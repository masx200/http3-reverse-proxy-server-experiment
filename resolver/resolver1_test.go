package resolver

import (
	"log"
	"testing"

	"github.com/fanjindong/go-cache"
	dns_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/dns"
	"github.com/masx200/http3-reverse-proxy-server-experiment/generic"
	h3_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/h3"
	"github.com/miekg/dns"
)

var DnsCache generic.MapInterface[string, cache.ICache] = dns_experiment.NewMapImplementSynchronous[string, cache.ICache]()

func DoHTTP3Client(m *dns.Msg, s string) (r *dns.Msg, err error) {
	return h3_experiment.DoHTTP3Client(m, s)
}
func GetQueryCallbacks2() generic.MapInterface[string, func(m *dns.Msg) (r *dns.Msg, err error)] {
	return generic.MapImplementFromMap(map[string]func(m *dns.Msg) (r *dns.Msg, err error){"https://cloudflare-dns.com/dns-query": func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://cloudflare-dns.com/dns-query")
	}, "https://dns.alidns.com/dns-query": func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://dns.alidns.com/dns-query")
	}, "https://unfiltered.adguard-dns.com/dns-query": func(m *dns.Msg) (r *dns.Msg, err error) {
		return DoHTTP3Client(m, "https://unfiltered.adguard-dns.com/dns-query")
	}, "https://security.cloudflare-dns.com/dns-query": func(m *dns.Msg) (r *dns.Msg, err error) {
		return DoHTTP3Client(m, "https://security.cloudflare-dns.com/dns-query")
	}})
}

func DohClient(m *dns.Msg, s string) (r *dns.Msg, err error) {
	return dns_experiment.DohClient(m, s)
}

func TestResolver77(t *testing.T) {
	x := "hello-word-worker-cloudflare.masx200.workers.dev"
	results, err := dns_experiment.DnsResolverMultipleServers(x, generic.MapImplementFromMap(map[string]func(m *dns.Msg) (r *dns.Msg, err error){"https://cloudflare-dns.com/dns-query": func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://cloudflare-dns.com/dns-query")
	}}), func(dro *dns_experiment.DnsResolverOptions) {
		dro.DnsCache = DnsCache
	})

	if err != nil {
		log.Printf("Error: %v\n", err)
		t.Error(err)
		return
	}

	for _, result := range results {
		log.Println(x, result)
	}
}
func TestResolver2(t *testing.T) {
	x := "nextjs-doh-reverse-proxy.onrender.com"
	results, err := dns_experiment.DnsResolverMultipleServers(x, generic.MapImplementFromMap(map[string]func(m *dns.Msg) (r *dns.Msg, err error){"https://cloudflare-dns.com/dns-query": func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://cloudflare-dns.com/dns-query")
	}}), func(dro *dns_experiment.DnsResolverOptions) {
		dro.DnsCache = DnsCache
	})

	if err != nil {
		log.Printf("Error: %v\n", err)
		t.Error(err)
		return
	}

	for _, result := range results {
		log.Println(x, result)
	}
}
func TestResolver3(t *testing.T) {
	x := "www.bilibili.com"
	results, err := dns_experiment.DnsResolverMultipleServers(x, generic.MapImplementFromMap(map[string]func(m *dns.Msg) (r *dns.Msg, err error){"https://cloudflare-dns.com/dns-query": func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://cloudflare-dns.com/dns-query")
	}}), func(dro *dns_experiment.DnsResolverOptions) {
		dro.DnsCache = DnsCache
	})

	if err != nil {
		log.Printf("Error: %v\n", err)
		t.Error(err)
		return
	}

	for _, result := range results {
		log.Println(x, result)
	}
}
func TestResolver4(t *testing.T) {
	x := "www.bilibili.com"
	results, err := dns_experiment.DnsResolverMultipleServers(x,
		generic.MapImplementFromMap(map[string]func(m *dns.Msg) (r *dns.Msg, err error){"https://cloudflare-dns.com/dns-query": func(m *dns.Msg) (r *dns.Msg, err error) {
			return DohClient(m, "https://cloudflare-dns.com/dns-query")
		}, "https://dns.alidns.com/dns-query": func(m *dns.Msg) (r *dns.Msg, err error) {
			return DohClient(m, "https://dns.alidns.com/dns-query")
		}}), func(dro *dns_experiment.DnsResolverOptions) {
			dro.DnsCache = DnsCache
		})
	if err != nil {
		log.Printf("Error: %v\n", err)
		t.Error(err)
		return
	}

	for _, result := range results {
		log.Println(x, result)
	}
}
func TestResolverMultipleServers77(t *testing.T) {
	x := "hello-word-worker-cloudflare.masx200.workers.dev"
	results, err := dns_experiment.DnsResolverMultipleServers(x,
		generic.MapImplementFromMap(map[string]func(m *dns.Msg) (r *dns.Msg, err error){"https://cloudflare-dns.com/dns-query": func(m *dns.Msg) (r *dns.Msg, err error) {
			return DohClient(m, "https://cloudflare-dns.com/dns-query")
		}, "https://dns.alidns.com/dns-query": func(m *dns.Msg) (r *dns.Msg, err error) {
			return DohClient(m, "https://dns.alidns.com/dns-query")
		}}), func(dro *dns_experiment.DnsResolverOptions) {
			dro.DnsCache = DnsCache
		})

	if err != nil {
		log.Printf("Error: %v\n", err)
		t.Error(err)
		return
	}

	for _, result := range results {
		log.Println(x, result)
	}
}
func TestResolver4224(t *testing.T) {
	x := "www.ithome.com"
	results, err := dns_experiment.DnsResolverMultipleServers(x, GetQueryCallbacks2(), func(dro *dns_experiment.DnsResolverOptions) {
		dro.DnsCache = DnsCache
	})

	if err != nil {
		log.Printf("Error: %v\n", err)
		t.Error(err)
		return
	}

	for _, result := range results {
		log.Println(x, result)
	}
}
func TestResolver_www_github(t *testing.T) {
	x := "www.github.com"
	results, err := dns_experiment.DnsResolverMultipleServers(x, GetQueryCallbacks2(), func(dro *dns_experiment.DnsResolverOptions) {
		dro.DnsCache = DnsCache
	})

	if err != nil {
		log.Printf("Error: %v\n", err)
		t.Error(err)
		return
	}

	for _, result := range results {
		log.Println(x, result)
	}
}
func TestResolver_github(t *testing.T) {
	x := "github.com"
	results, err := dns_experiment.DnsResolverMultipleServers(x, GetQueryCallbacks2(), func(dro *dns_experiment.DnsResolverOptions) {
		dro.DnsCache = DnsCache
	})

	if err != nil {
		log.Printf("Error: %v\n", err)
		t.Error(err)
		return
	}

	for _, result := range results {
		log.Println(x, result)
	}
	x2 := "github.com"
	results2, err := dns_experiment.DnsResolverMultipleServers(x2, GetQueryCallbacks2(), func(dro *dns_experiment.DnsResolverOptions) {
		dro.DnsCache = DnsCache
	})

	if err != nil {
		t.Error(err)
		log.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results2 {
		log.Println(x, result)
	}
}
