package dns

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/miekg/dns"
	doq "github.com/tantalor93/doq-go/doq"
)

func TestDOQ(t *testing.T) {
	// 创建一个新的 DoQ 客户端
	x := "family.adguard-dns.com:853"
	doqServer := "quic://" + x
	client := doq.NewClient(x, doq.Options{})
	// if err != nil {
	// 	panic(err)
	// }

	domain := "production.hello-word-worker-cloudflare.masx200.workers.dev"

	// 查询 A 记录
	qA := dns.Msg{}
	qA.SetQuestion(domain+".", dns.TypeA)
	respA, err := client.Send(context.Background(), &qA)
	if err != nil {
		fmt.Println("Error querying A record:", err)
		t.Fatal(err)
	} else {
		fmt.Println("A Record Response:", respA.String())
	}
	if respA.Rcode != dns.RcodeSuccess {
		log.Println(dns.RcodeToString[respA.Rcode])

		t.Fatal(fmt.Errorf("dns server %s response error not success", doqServer))
	}
	if len(respA.Answer) == 0 {
		log.Println(doqServer + "-No A records found")
		t.Fatal(fmt.Errorf(
			"dns server  response error No A records found",
		))
	}
	// 查询 AAAA 记录
	qAAAA := dns.Msg{}
	qAAAA.SetQuestion(domain+".", dns.TypeAAAA)
	respAAAA, err := client.Send(context.Background(), &qAAAA)
	if err != nil {
		fmt.Println("Error querying AAAA record:", err)
		t.Fatal(err)
	} else {
		fmt.Println("AAAA Record Response:", respAAAA.String())
	}
	if respAAAA.Rcode != dns.RcodeSuccess {
		log.Println(dns.RcodeToString[qAAAA.Rcode])

		t.Fatal(fmt.Errorf("dns server %s response error not success", doqServer))
	}
	if len(respAAAA.Answer) == 0 {
		log.Println(doqServer + "-No AAAA records found")
		t.Fatal(fmt.Errorf(
			"dns server  response error No AAAA records found",
		))
	}
	// 查询 HTTPS 记录（HTTPS 相关）
	// 注意：这里假设服务器支持并返回 HTTPS 记录
	qHTTPS := dns.Msg{}
	qHTTPS.SetQuestion(domain+".", dns.TypeHTTPS)
	respHTTPS, err := client.Send(context.Background(), &qHTTPS)
	if err != nil {
		fmt.Println("Error querying HTTPS record:", err)
		t.Fatal(err)
	} else {
		fmt.Println("HTTPS Record Response:", respHTTPS.String())
	}
	if respHTTPS.Rcode != dns.RcodeSuccess {
		log.Println(dns.RcodeToString[respHTTPS.Rcode])

		t.Fatal(fmt.Errorf("dns server %s response error not success", doqServer))
	}
	if len(respHTTPS.Answer) == 0 {
		log.Println(doqServer + "-No HTTPS records found")
		t.Fatal(fmt.Errorf(
			"dns server  response error No HTTPS records found",
		))
	}
}
