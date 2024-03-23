package h3

import "testing"

func TestCheckHttp3ViaDNS(t *testing.T) {

	DOHServer := "https://deno-dns-over-https-server-5ehq9rg3chgf.deno.dev/dns-query"

	domain := "production.hello-word-worker-cloudflare.masx200.workers.dev" //"quic.nginx.org" //
	port := "443"

	supportsH3, err := CheckHttp3ViaDNS(domain, port, DOHServer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !supportsH3 {
		t.Errorf("expected H3 support, but got false")
	}
}
func TestCheckHttp3ViaHttp2(t *testing.T) {

	domain := "quic.nginx.org" //"quic.nginx.org" //
	port := "443"

	supportsH3, err := CheckHttp3ViaHttp2(domain, port)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !supportsH3 {
		t.Errorf("expected H3 support, but got false")
	}
}
