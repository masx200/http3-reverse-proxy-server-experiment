package dns

import (
	"fmt"
	"testing"

	"github.com/miekg/dns"
)

func TestResolver(t *testing.T) {
	x := "hello-word-worker-cloudflare.masx200.workers.dev"
	results, err := DnsResolver(func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://cloudflare-dns.com/dns-query")
	}, x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
func TestResolver2(t *testing.T) {
	x := "nextjs-doh-reverse-proxy.onrender.com"
	results, err := DnsResolver(func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://cloudflare-dns.com/dns-query")
	}, x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
func TestResolver3(t *testing.T) {
	x := "www.bilibili.com"
	results, err := DnsResolver(func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://cloudflare-dns.com/dns-query")
	}, x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
func TestResolver4(t *testing.T) {
	x := "www.bilibili.com"
	results, err := DnsResolverMultipleServers([]func(m *dns.Msg) (r *dns.Msg, err error){func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://cloudflare-dns.com/dns-query")
	}, func(m *dns.Msg) (r *dns.Msg, err error) {
		return DohClient(m, "https://dns.alidns.com/dns-query")
	}}, x)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Println(x, result)
	}
}
