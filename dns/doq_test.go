package dns

import (
	"context"
	"fmt"
	"testing"

	"github.com/miekg/dns"
	doq "github.com/tantalor93/doq-go/doq"
)

func TestDOQ(t *testing.T) {
	// 创建一个新的 DoQ 客户端
	client := doq.NewClient("family.adguard-dns.com:853", doq.Options{})
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
}
