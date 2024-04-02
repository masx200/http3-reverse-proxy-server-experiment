package resolver

import (
	"fmt"
	"testing"

	dns_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/dns"
	"github.com/miekg/dns"
)

func TestResolver9(t *testing.T) {
	x := "hello-word-worker-cloudflare.masx200.workers.dev"
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks2(), x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
func TestResolver28(t *testing.T) {
	x := "nextjs-doh-reverse-proxy.onrender.com"
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks2(), x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
func TestResolver37(t *testing.T) {
	x := "www.bilibili.com"
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks2(), x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
func GetQueryCallbacks14() []func(m *dns.Msg) (r *dns.Msg, err error) {
	return []func(m *dns.Msg) (r *dns.Msg, err error){func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://cloudflare-dns.com/dns-query")
	}, func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://dns.alidns.com/dns-query")
	}, func(m *dns.Msg) (r *dns.Msg, err error) {
		return DoHTTP3Client(m, "https://doh-cache-worker-cf.masx200.workers.dev/dns-query")
	}, func(m *dns.Msg) (r *dns.Msg, err error) {
		return DoHTTP3Client(m, "https://dns.alidns.com/dns-query")
	}, func(m *dns.Msg) (r *dns.Msg, err error) {
		return DoQClient(m, "quic://dns.alidns.com")
	}, func(m *dns.Msg) (r *dns.Msg, err error) {
		return DoQClient(m, "quic://family.adguard-dns.com")
	}, func(m *dns.Msg) (r *dns.Msg, err error) {
		return DoTClient(m, "tls://dot.pub")
	}}
}

func DoTClient(m *dns.Msg, s string) (r *dns.Msg, err error) {
	return dns_experiment.DoTClient(m, s)
}

func DoQClient(m *dns.Msg, s string) (r *dns.Msg, err error) {
	return dns_experiment.DoQClient(m, s)
}
func TestResolver42(t *testing.T) {
	x := "www.bilibili.com"
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks2(), x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
func TestResolverMultipleServers2(t *testing.T) {
	x := "hello-word-worker-cloudflare.masx200.workers.dev"
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks2(), x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
func TestResolver424(t *testing.T) {
	x := "www.bilibili.com"
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks14(), x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
func TestResolverMultipleServers24(t *testing.T) {
	x := "hello-word-worker-cloudflare.masx200.workers.dev"
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks14(), x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
func TestResolverMultipleServers234(t *testing.T) {
	x := "www.fastly.com"
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks14(), x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
