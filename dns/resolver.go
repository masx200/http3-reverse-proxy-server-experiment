package dns

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/miekg/dns"
)

// DnsResolverOptions 是DNS解析器的配置选项。
type DnsResolverOptions struct {
	QueryCallback func(m *dns.Msg) (r *dns.Msg, err error) // QueryCallback 是一个回调函数，用于自定义DNS查询逻辑。接收一个dns.Msg类型的参数，返回一个dns.Msg类型和error类型的值。
	Domain        string                                   // Domain 是需要进行DNS解析的域名。

	HttpsPort int // HttpsPort 是HTTPS服务监听的端口号。
}

// DnsResolver 是一个用于解析特定域名下多种类型记录的函数，例如A记录、AAAA记录和HTTPS记录。
// queryCallback 是一个回调函数，用于执行DNS查询并返回结果。
// domain 是需要查询的域名。
// optionsCallBacks 是一个可选参数列表，用于修改查询选项。
// 返回解析到的地址列表和可能发生的错误。
func DnsResolver(queryCallback func(m *dns.Msg) (r *dns.Msg, err error), domain string, optionsCallBacks ...func(*DnsResolverOptions)) ([]string, error) {

	var options = &DnsResolverOptions{QueryCallback: queryCallback, Domain: domain, HttpsPort: 443}

	for _, optionsCallBack := range optionsCallBacks {
		optionsCallBack(options)
	}
	var resultsMutex sync.Mutex
	var results []string
	var wg sync.WaitGroup
	var tasks = []func(){
		func() {
			defer wg.Done()

			res, err := resolve(options, dns.TypeA)
			if err != nil {
				fmt.Printf("Error querying A record for %s: %v\n", options.Domain, err)
				return
			}
			resultsMutex.Lock()
			defer resultsMutex.Unlock()
			results = append(results, res...)

		}, func() {
			defer wg.Done()
			res, err := resolve(options, dns.TypeAAAA)
			if err != nil {
				fmt.Printf("Error querying AAAA record for %s: %v\n", options.Domain, err)
				return
			}
			resultsMutex.Lock()
			defer resultsMutex.Unlock()
			results = append(results, res...)
		}, func() {
			defer wg.Done()
			res, err := resolve(options, dns.TypeHTTPS)
			if err != nil {
				fmt.Printf("Error querying HTTPS record for %s: %v\n", options.Domain, err)
				return
			}
			resultsMutex.Lock()
			defer resultsMutex.Unlock()
			results = append(results, res...)
		},
	}
	wg.Add(len(tasks))
	for _, task := range tasks {
		go task()
	}

	wg.Wait()
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for %s", options.Domain)
	}
	return removeDuplicates(results), nil

} // resolve 是一个用于解析特定域名下指定类型记录的函数。
// options: 指定DNS解析器的选项，包含域名、端口和其他配置。
// recordType: 指定需要查询的记录类型（如A记录、AAAA记录等）。
// 返回值为解析到的记录值字符串数组和可能发生的错误。
func resolve(options *DnsResolverOptions, recordType uint16) ([]string, error) {
	m := &dns.Msg{}
	if recordType == dns.TypeHTTPS && options.HttpsPort != 443 {

		m.SetQuestion(fmt.Sprintf("_%s._https.", fmt.Sprint(options.HttpsPort))+dns.Fqdn(options.Domain), recordType)
	} else {
		m.SetQuestion(dns.Fqdn(options.Domain), recordType)
	}

	fmt.Println(m)
	resp, err := options.QueryCallback(m)
	if err != nil {
		return nil, err
	}
	if resp.Rcode != dns.RcodeSuccess {
		log.Println(dns.RcodeToString[resp.Rcode])
		return nil, fmt.Errorf("dns server  response error not success:" + m.Question[0].String())
	}
	if len(resp.Answer) == 0 {
		log.Println("dns server-No  records found:" + m.Question[0].String())
		return nil, fmt.Errorf(
			"dns server  response error No  records found:" + m.Question[0].String(),
		)
	}
	fmt.Println(resp)
	var results []string
	for _, answer := range resp.Answer {
		switch record := answer.(type) {
		case *dns.A:
			results = append(results, (record.A.String()))
		case *dns.AAAA:
			results = append(results, (record.AAAA.String()))
		case *dns.HTTPS:
			{
				if record.Priority != 0 {
					for _, value := range record.Value {
						if value.Key().String() == "ipv4hint" {
							var addresses = strings.Split(value.String(), ",")
							results = append(results, addresses...)

						} else if value.Key().String() == "ipv6hint" {
							var addresses = strings.Split(value.String(), ",")
							results = append(results, addresses...)
						}
					}
				}
			}
		case *dns.CNAME:
			// results = append(results, fmt.Sprintf("CNAME: %s", record.Target))
			res, err := DnsResolver(options.QueryCallback, record.Target, func(dro *DnsResolverOptions) {
				dro.HttpsPort = options.HttpsPort
			})
			if err != nil {
				return nil, err
			}
			results = append(results, res...)
		}
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for %s", options.Domain)
	}
	return removeDuplicates(results), nil
} // removeDuplicates 函数用于移除一个可比较类型切片中的重复元素。
// 参数 arr 是待处理的切片，函数返回一个不包含重复元素的新切片。
// [T comparable] 使用了泛型 T，限制 T 必须是可比较的类型。
func removeDuplicates[T comparable](arr []T) []T {
	seen := make(map[T]bool)
	var result []T

	for _, value := range arr {
		if _, ok := seen[value]; !ok {
			seen[value] = true
			result = append(result, value)
		}
	}

	return result
}
