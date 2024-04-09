package main

import (
	"fmt"
	"os"
	"strings"

	h3_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/h3"
	"github.com/miekg/dns"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("usage:", os.Args[0], "domain dnstype dohurl")
		return
	}

	var domain = os.Args[1]
	var dnstype = os.Args[2]
	var dohurl = os.Args[3]
	fmt.Println("domain:", domain, "dnstype:", dnstype, "dohurl:", dohurl)

	for _, d := range strings.Split(domain, ",") {
		for _, t := range strings.Split(dnstype, ",") {
			fmt.Println("domain:", d, "dnstype:", t, "dohurl:", dohurl)

			var msg = &dns.Msg{}
			msg.SetQuestion(d+".", dns.StringToType[t])
			fmt.Println(msg.String())

			res, err := h3_experiment.DoHTTP3Client(msg, dohurl)
			if err != nil {
				fmt.Println(err)
				return

			}
			fmt.Println(res.String())
		}
	}

}
