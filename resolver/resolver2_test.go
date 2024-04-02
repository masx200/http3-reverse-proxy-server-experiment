package resolver

import (
	"fmt"
	dns_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/dns"
	"github.com/miekg/dns"
	"testing"
)

func TestResolver9(t *testing.T) {
	x := "hello-word-worker-cloudflare.masx200.workers.dev"
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks(), x)

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
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks(), x)

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
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks(), x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
func GetQueryCallbacks1() []func(m *dns.Msg) (r *dns.Msg, err error) {
	return []func(m *dns.Msg) (r *dns.Msg, err error){func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://cloudflare-dns.com/dns-query")
	}, func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://dns.alidns.com/dns-query")
	}, func(m *dns.Msg) (r *dns.Msg, err error) {
		return DoHTTP3Client(m, "https://doh-cache-worker-cf.masx200.workers.dev/dns-query")
	}, func(m *dns.Msg) (r *dns.Msg, err error) {
		return DoHTTP3Client(m, "https://dns.alidns.com/dns-query")
	}}
}
func TestResolver42(t *testing.T) {
	x := "www.bilibili.com"
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks(), x)

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
	results, err := dns_experiment.DnsResolverMultipleServers(GetQueryCallbacks(), x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
