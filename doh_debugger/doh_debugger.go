package main

import (
	"flag"
	"log"
	"strings"
	"sync"

	dns_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/dns"
	"github.com/miekg/dns"
)

func main() {

	domain := flag.String("domain", "", "指定要查询的域名")
	dohurl := flag.String("dohurl", "", "指定DoH(DNS over HTTPS)服务的URL")

	// 定义可选的命令行标志
	dnstype := flag.String("dnstype", "A,AAAA", "指定DNS查询类型，默认为A和AAAA记录")
	dohip := flag.String("dohip", "", "指定DoH服务的IP地址（可选）")

	// 解析命令行参数
	flag.Parse()

	// 必需参数检查
	if *domain == "" || *dohurl == "" {
		log.Println("错误：必须提供-domain和-dohurl参数")
		flag.Usage()
		return
	}
	if *dohip != "" {
		dohnslookup(*domain, *dnstype, *dohurl, *dohip)
	} else {
		dohnslookup(*domain, *dnstype, *dohurl)
	}
}

func dohnslookup(domain string, dnstype string, dohurl string, dohip ...string) {
	log.Println("domain:", domain, "dnstype:", dnstype, "dohurl:", dohurl)
	var wg sync.WaitGroup
	for _, d := range strings.Split(domain, ",") {
		for _, t := range strings.Split(dnstype, ",") {
			wg.Add(1)
			go func(d string, t string) {
				defer wg.Done()
				log.Println("domain:", d, "dnstype:", t, "dohurl:", dohurl)
				var msg = &dns.Msg{}
				msg.SetQuestion(d+".", dns.StringToType[t])
				log.Println(msg.String())

				res, err := dns_experiment.DohClient(msg, dohurl, dohip...)
				if err != nil {
					log.Println(err)
					return

				}
				log.Println(res.String())
			}(d, t)

		}
	}
	wg.Wait()
}
