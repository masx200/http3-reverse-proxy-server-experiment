package h3

import (
	"log"
	"testing"
)

func TestCheckHttp3ViaDNS(t *testing.T) {

	DOHServer := "https://dns.dnswarden.com/uncensored"

	domain := "production.hello-word-worker-cloudflare.masx200.workers.dev"
	port := "443"

	supportsH3, err := CheckHttp3ViaDNS(domain, port, DOHServer)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !supportsH3 {
		t.Errorf("expected H3 support, but got false")
	}
	log.Println("H3 support:", supportsH3, domain, port)
}
func TestCheckHttp3ViaHttp2(t *testing.T) {
	port := "443"
	domain := "quic.nginx.org"

	supportsH3, err := CheckHttp3ViaHttp2(domain, port)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !supportsH3 {
		t.Errorf("expected H3 support, but got false")
	}
	log.Println("H3 support:", supportsH3, domain, port)
}
