package dns

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"crypto/sha512"
	"encoding/hex"

	"github.com/fanjindong/go-cache"
	"github.com/masx200/http3-reverse-proxy-server-experiment/generic"
	"github.com/miekg/dns"
)

// "github.com/masx200/http3-reverse-proxy-server-experiment/generic"

// Sha512 计算给定字节切片的SHA-512哈希值，并以字符串形式返回
func Sha512(input []byte) string {
	hash := sha512.Sum512(input)
	return hex.EncodeToString(hash[:])
}

// DnsResolverMultipleServers 使用多个DNS查询回调函数来解析给定域名，并返回解析结果的去重列表。
// queryCallbacks: 一个包含多个DNS查询回调函数的切片，每个函数尝试解析指定的域名。
// domain: 需要解析的域名。
// optionsCallBacks: 可选的一组函数，用于定制DNS解析器的选项。
// 返回值: 解析到的IP地址字符串切片（去重后），如果没有任何解析结果，则返回错误。
func DnsResolverMultipleServers(domain string, queryCallbacks generic.MapInterface[string, func(m *dns.Msg) (r *dns.Msg, err error)], optionsCallBacks ...func(*DnsResolverOptions)) ([]string, error) {
	// c := cache.NewMemCache()
	var options = &DnsResolverOptions{
		HttpsPort:     443,
		QueryCallback: queryCallbacks,
		Domain:        domain,
		DnsCache:      generic.NewMapImplement[string, cache.ICache](),
	}
	for _, optionsCallBack := range optionsCallBacks {
		optionsCallBack(options)
	}
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	var results []string
	if (queryCallbacks).Size() == 0 {
		return nil, fmt.Errorf("no query callbacks provided")
	}
	queryCallbacks.ForEach(func(queryCallback func(m *dns.Msg) (r *dns.Msg, err error), s string, mi generic.MapInterface[string, func(m *dns.Msg) (r *dns.Msg, err error)]) {
		wg.Add(1)
		go func(queryCallback func(m *dns.Msg) (r *dns.Msg, err error)) {
			defer wg.Done()
			res, err := DnsResolver(func(m *dns.Msg) (*dns.Msg, error) {
				var a, b = options.DnsCache.Get(s)
				if !(a != nil && b) {
					a = cache.NewMemCache()
					options.DnsCache.Set(s, a)
				}
				var buffer, err = m.Pack()
				if err != nil {
					log.Println(s, err)
					return nil, err
				}
				var hash = Sha512(buffer)

				var c, d = a.Get(hash)

				if c != nil && d {
					log.Println(s, "cache hit", hash)
					return c.(*dns.Msg), nil
				}
				result, err := queryCallback(m)
				if err != nil {
					log.Println(s, err)
					return nil, err
				}
				a.Set(hash, result)
				log.Println(s, "cache miss", hash)
				return result, nil
			}, domain, options.HttpsPort)
			if err != nil {
				fmt.Printf("Error resolving domain %s: %v\n", domain, err)
				return
			}
			resultsMutex.Lock()
			defer resultsMutex.Unlock()
			results = append(results, res...)
		}(queryCallback)

	})

	wg.Wait()
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for %s", domain)
	}
	return removeDuplicates(results), nil
}

// DnsResolverOptions 是DNS解析器的配置选项。
type DnsResolverOptions struct {
	QueryCallback generic.MapInterface[string, func(m *dns.Msg) (r *dns.Msg, err error)] // QueryCallback 是一个回调函数，用于自定义DNS查询逻辑。接收一个dns.Msg类型的参数，返回一个dns.Msg类型和error类型的值。
	Domain        string                                                                 // Domain 是需要进行DNS解析的域名。
	DnsCache      generic.MapInterface[string, cache.ICache]
	HttpsPort     int // HttpsPort 是HTTPS服务监听的端口号。
}

// DnsResolver 是一个用于解析特定域名下多种类型记录的函数，例如A记录、AAAA记录和HTTPS记录。
// queryCallback 是一个回调函数，用于执行DNS查询并返回结果。
// domain 是需要查询的域名。
// optionsCallBacks 是一个可选参数列表，用于修改查询选项。
// 返回解析到的地址列表和可能发生的错误。
func DnsResolver(queryCallback func(m *dns.Msg) (r *dns.Msg, err error), domain string, HttpsPort int) ([]string, error) {

	var resultsMutex sync.Mutex
	var results []string
	var wg sync.WaitGroup
	var tasks = []func(){
		func() {
			defer wg.Done()

			res, err := resolve(dns.TypeA, queryCallback, domain, HttpsPort)
			if err != nil {
				fmt.Printf("Error querying A record for %s: %v\n", domain, err)
				return
			}
			resultsMutex.Lock()
			defer resultsMutex.Unlock()
			results = append(results, res...)

		}, func() {
			defer wg.Done()
			res, err := resolve(dns.TypeAAAA, queryCallback, domain, HttpsPort)
			if err != nil {
				fmt.Printf("Error querying AAAA record for %s: %v\n", domain, err)
				return
			}
			resultsMutex.Lock()
			defer resultsMutex.Unlock()
			results = append(results, res...)
		}, func() {
			defer wg.Done()
			res, err := resolve(dns.TypeHTTPS, queryCallback, domain, HttpsPort)
			if err != nil {
				fmt.Printf("Error querying HTTPS record for %s: %v\n", domain, err)
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
		return nil, fmt.Errorf("no results found for %s", domain)
	}
	return removeDuplicates(results), nil

} // resolve 是一个用于解析特定域名下指定类型记录的函数。
// options: 指定DNS解析器的选项，包含域名、端口和其他配置。
// recordType: 指定需要查询的记录类型（如A记录、AAAA记录等）。
// 返回值为解析到的记录值字符串数组和可能发生的错误。
func resolve(recordType uint16, QueryCallback func(m *dns.Msg) (r *dns.Msg, err error), domain string, HttpsPort int) ([]string, error) {
	m := &dns.Msg{}
	if recordType == dns.TypeHTTPS && HttpsPort != 443 {

		m.SetQuestion(fmt.Sprintf("_%s._https.", fmt.Sprint(HttpsPort))+dns.Fqdn(domain), recordType)
	} else {
		m.SetQuestion(dns.Fqdn(domain), recordType)
	}

	fmt.Println(m)
	resp, err := QueryCallback(m)
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
			res, err := DnsResolver(QueryCallback, record.Target, HttpsPort)
			if err != nil {
				return nil, err
			}
			results = append(results, res...)
		}
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for %s", domain)
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
