package h3

import "testing"

func TestCheckHttp3ViaDNS(t *testing.T) {
	domain := "production.hello-word-worker-cloudflare.masx200.workers.dev"
	port := "443"
	DOHServer := "https://deno-dns-over-https-server-5ehq9rg3chgf.deno.dev/dns-query"

	supportsH3, err := CheckHttp3ViaDNS(domain, port, DOHServer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !supportsH3 {
		t.Errorf("expected H3 support, but got false")
	}
}
