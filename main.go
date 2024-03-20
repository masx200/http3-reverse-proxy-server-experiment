package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// checkUpstreamHealth 检查上游服务的健康状态
// url: 要检查的服务的URL地址
// RoundTrip: 用于发送HTTP请求的函数，模拟HTTP客户端的行为
// 返回值: 返回一个布尔值，表示上游服务的健康状态，true为健康，false为不健康

func checkUpstreamHealth(url string, RoundTrip func(*http.Request) (*http.Response, error)) bool {

	// 发送HEAD请求并检查返回的状态码
	statusCode, err := sendHeadRequestAndCheckStatus(url, RoundTrip)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}

	// 打印状态码信息
	fmt.Printf("health check Status code: %d\n", statusCode)

	// 根据状态码判断上游服务的健康状态
	if statusCode < 500 {
		fmt.Println("Status code is less than 500.")
		return true
	} else {
		fmt.Println("Status code is 500 or greater.")
	}
	return false
}

// RoundTripTransport 是一个实现了 http.RoundTripper 接口的类型，
// 允许自定义HTTP请求的传输行为。
type RoundTripTransport struct {
	roundTrip func(*http.Request) (*http.Response, error) // roundTrip 是一个函数，它执行HTTP请求的传输，并返回响应和可能的错误。
}

// RoundTrip 是 http.RoundTripper 接口要求的方法，用于执行HTTP请求。
// 它简单地调用了结构体中的 roundTrip 函数，传递给它一个HTTP请求，并返回该请求的响应及可能的错误。
func (r *RoundTripTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return r.roundTrip(req)
}

// sendHeadRequestAndCheckStatus 发送一个HEAD请求并检查状态码。
// url: 请求的目标URL。
// RoundTrip: 自定义的HTTP.RoundTripper函数，用于发送请求。
// 返回值: 请求的状态码和可能出现的错误。
func sendHeadRequestAndCheckStatus(url string, RoundTrip func(*http.Request) (*http.Response, error)) (int, error) {
	client := &http.Client{}

	client.Transport = &RoundTripTransport{
		roundTrip: RoundTrip,
	}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}
	PrintRequest(req)
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	PrintResponse(resp)
	return resp.StatusCode, nil
}

// 主程序入口
func main() {

	var httpsPort = 18443
	var httpPort = 18080
	var upStreamServerSchemeAndHostOfName map[string]Pair[string, string] = map[string]Pair[string, string]{}
	r := gin.Default()
	r.Use(Forwarded(), LoopDetect())
	// 定义上游服务器地址
	var upstreamServers = []string{"https://quic.nginx.org/", "https://www.baidu.com/", "https://www.so.com/", "https://hello-world-deno-deploy.deno.dev/", "https://production.hello-word-worker.masx200.workers.dev/"}
	//打印上游
	fmt.Println("Upstream servers:")
	for _, server := range upstreamServers {
		fmt.Println(server)
	}
	//过期时间毫秒
	var maxAge = 30 * 1000
	var expires = int64(0)
	var upstreamServerOfName = map[string]string{}
	var proxyServers map[string]func(*http.Request) (*http.Response, error) = map[string]func(*http.Request) (*http.Response, error){}

	for _, urlString := range upstreamServers {
		upstreamURL, err := url.Parse(urlString)
		if err != nil {
			log.Fatalf("Failed to parse upstream server URL: %v", err)
		}
		if upstreamURL.Path != "/" && upstreamURL.Path != "" {
			log.Fatalf("upstreamServer Path must be / or empty")
		}
		proxy, err := createReverseProxy(urlString)

		if err != nil {
			log.Fatal(err)
		}
		proxyServers[urlString] = func(req *http.Request) (*http.Response, error) {
			return proxy.Transport.RoundTrip(req)
		}
		upstreamServerOfName[urlString] = urlString
		upStreamServerSchemeAndHostOfName[urlString] = Pair[string, string]{upstreamURL.Scheme, upstreamURL.Host}
	}
	// 启动反向代理服务器
	var healthyUpstream = proxyServers
	var transportsUpstream = proxyServers

	var getHealthyProxyServers = func() map[string]func(*http.Request) (*http.Response, error) {
		return refreshHealthyUpStreams(func() int64 { return expires }, func() map[string]func(*http.Request) (*http.Response, error) { return healthyUpstream }, transportsUpstream, upstreamServerOfName, maxAge, func(i int64) { expires = i }, func(transportsUpstream map[string]func(*http.Request) (*http.Response, error)) {
			healthyUpstream = transportsUpstream
		})

	}
	r.Any("/*path", func(c *gin.Context) {
		req := c.Request
		if req.TLS != nil {
			req.URL.Scheme = "https"
		} else {
			req.URL.Scheme = "http"
		}

		req.URL.Host = req.Host
		PrintRequest(req) // 打印请求信息

		// 使用随机负载均衡策略选择一个健康状态的传输函数，并执行请求
		var resp, err = RandomLoadBalancer(getHealthyProxyServers(), req, upStreamServerSchemeAndHostOfName)

		if err != nil {
			fmt.Println("ERROR:", err) // 打印错误信息
			c.AbortWithStatus(http.StatusBadGateway)
			return
		} else {
			PrintResponse(resp) // 打印响应信息
		}
		for k, vv := range resp.Header {
			for _, v := range vv {
				c.Header(k, v)
			}
		}
		c.Status(resp.StatusCode)
		defer resp.Body.Close()
		bufio.NewReader(resp.Body).WriteTo(c.Writer)

	})
	server := &http.Server{
		Addr: ":" + strconv.Itoa(httpsPort),
		Handler: &LoadBalanceHandler{
			engine: r,
		},
	}

	var hostname = "0.0.0.0"
	go func() {
		listener, err := net.Listen("tcp", hostname+":"+fmt.Sprint(httpPort))
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
		log.Printf("http reverse proxy server started on port %s", listener.Addr())

		// 设置自定义处理器
		http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			r.Handler().ServeHTTP(w, req)
		})

		// 开始服务
		err = http.Serve(listener, nil)
		if err != nil {
			log.Fatal("Serve: ", err)
		}
	}()
	log.Printf("Starting http reverse proxy server on :" + strconv.Itoa(httpsPort))
	x := server.ListenAndServeTLS("cert.crt", "key.pem")
	log.Fatal(x)
}

func Forwarded() gin.HandlerFunc {
	return func(c *gin.Context) {
		var clienthost = c.RemoteIP()
		var address = c.Request.Host
		if address == "" {
			address = c.Request.URL.Host
		}
		var proto = "http"

		if c.Request.TLS != nil {
			proto = "https"
		}
		forwarded := fmt.Sprintf(
			"for=%s;by=%s;host=%s;proto=%s",
			clienthost,     // 代理自己的标识或IP地址
			c.Request.Host, // 代理的标识
			address,        // 原始请求的目标主机名
			proto,          // 或者 "https" 根据实际协议
		)
		c.Request.Header.Add("Forwarded", forwarded)
		c.Next()
	}
}
func LoopDetect() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// LoadBalanceHandler 是一个用于负载均衡处理的结构体。
type LoadBalanceHandler struct {
	engine *gin.Engine // engine 是一个Gin框架的引擎实例，用于处理HTTP请求。
}

// ServeHTTP 是一个实现http.Handler接口的方法，用于处理HTTP请求。
// 参数w是用于向客户端发送响应的http.ResponseWriter，
// 参数req是客户端发来的HTTP请求。
func (h *LoadBalanceHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.engine.Handler().ServeHTTP(w, req) // 调用Gin引擎的Handler方法处理HTTP请求。
}
func createReverseProxy(upstreamServer string) (*httputil.ReverseProxy, error) {
	// 解析上游服务器URL，确保其路径为根路径或为空
	upstreamURL, err := url.Parse(upstreamServer)
	if err != nil {
		log.Fatalf("Failed to parse upstream server URL: %v", err)
	}
	if upstreamURL.Path != "/" && upstreamURL.Path != "" {
		log.Fatalf("upstreamServer Path must be / or empty")
	}

	// 初始化HTTP/3客户端
	http3Client := &http.Client{
		Transport: &http3.RoundTripper{
			TLSClientConfig: &tls.Config{},
			QuicConfig:      &quic.Config{},
		},
	}
	// 初始化HTTP/2客户端，使用默认传输
	http2Client := &http.Client{
		Transport: http.DefaultTransport,
	}

	// 定义上游服务器的传输方式，包括HTTP/3和HTTP/2
	var transportsUpstream = map[string]func(*http.Request) (*http.Response, error){
		"http3": func(req *http.Request) (*http.Response, error) { return http3Client.Transport.RoundTrip(req) },

		"http2": func(req *http.Request) (*http.Response, error) { return http2Client.Transport.RoundTrip(req) },
	}
	var upstreamServerOfName = map[string]string{}

	for k := range transportsUpstream {
		upstreamServerOfName[k] = upstreamServer
	}
	// 设置健康检查的超时时间毫秒
	var maxAge = 30 * 1000
	var expires = int64(0)
	var healthyUpstream = transportsUpstream

	// 自定义负载均衡的传输器，根据上游服务器的健康状况选择传输方式
	customTransport := &customRoundTripperLoadBalancer{
		upstreamURL: upstreamURL,
		getTransportHealthy: func() map[string]func(*http.Request) (*http.Response, error) {
			// 对上游服务器进行健康检查，选择健康的传输方式
			return refreshHealthyUpStreams(func() int64 { return expires }, func() map[string]func(*http.Request) (*http.Response, error) { return healthyUpstream }, transportsUpstream, upstreamServerOfName, maxAge, func(i int64) { expires = i }, func(transportsUpstream map[string]func(*http.Request) (*http.Response, error)) {
				healthyUpstream = transportsUpstream
			})
		},
	}

	// 初始化反向代理
	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)

	// 设置反向代理的传输器为自定义的负载均衡传输器
	proxy.Transport = customTransport
	return proxy, err
}

// refreshHealthyUpStreams 刷新健康状态的上游服务器。
// 此函数会根据上游服务器的当前健康状态和过期时间来更新健康的上游服务器列表。
//
// 参数:
// - getExpires: 获取当前上游服务器列表的过期时间的函数。
// - healthyUpstream: 当前已知的健康上游服务器映射。
// - transportsUpstream: 所有可用的上游服务器传输函数映射。
// - upstreamServer: 上游服务器地址。
// - maxAge: 上游服务器列表的有效期（毫秒）。
// - setExpires: 设置上游服务器列表过期时间的函数。
//
// 返回值:
// - 返回更新后的健康上游服务器映射。
func refreshHealthyUpStreams(getExpires func() int64, getHealthyUpstream func() map[string]func(*http.Request) (*http.Response, error), transportsUpstream map[string]func(*http.Request) (*http.Response, error), upstreamServerOfName map[string]string, maxAge int, setExpires func(int64), setHealthyUpstream func(transportsUpstream map[string]func(*http.Request) (*http.Response, error))) map[string]func(*http.Request) (*http.Response, error) {
	// 检查当前上游服务器列表是否已过期。
	if getExpires() > time.Now().UnixMilli() {
		fmt.Println("不需要进行健康检查", "还剩余的时间毫秒", getExpires()-time.Now().UnixMilli())
		fmt.Println("healthyUpstream", getHealthyUpstream())
		return getHealthyUpstream()
	}

	// 在后台进行健康检查更新。
	go func() {
		var healthy = map[string]func(*http.Request) (*http.Response, error){}
		fmt.Println("需要进行健康检查", "已经过期的时间毫秒", -getExpires()+time.Now().UnixMilli())
		//需要并行检查
		// 遍历所有上游服务器进行健康检查。
		for key, roundTrip := range transportsUpstream {
			keyi0 := key
			roundTripi0 := roundTrip
			go func() {
				var upstreamServer = upstreamServerOfName[keyi0]
				//loop variable roundTrip captured by func literal loop closure
				if checkUpstreamHealth(upstreamServer, roundTripi0) {
					healthy[keyi0] = roundTripi0
					fmt.Println("健康检查成功", keyi0, upstreamServer)
				} else {
					fmt.Println("健康检查失败", keyi0, upstreamServer)
				}
			}()

		}

		// 根据健康检查结果更新健康上游服务器列表。
		if len(healthy) == 0 {
			setHealthyUpstream(transportsUpstream)
		} else {
			setHealthyUpstream(healthy)
			fmt.Println("healthyUpstream", getHealthyUpstream())
		}

		// 设置上游服务器列表的新过期时间。
		var expires = time.Now().UnixMilli() + int64(maxAge)
		setExpires(expires)
	}()

	// 返回当前的健康上游服务器列表。
	return getHealthyUpstream()
}

// customRoundTripperLoadBalancer 是一个自定义的负载均衡器，
// 用于在多个上游服务之间进行请求的负载均衡。
type customRoundTripperLoadBalancer struct {
	upstreamURL         *url.URL                                                      // upstreamURL 指定了上游服务的URL
	getTransportHealthy func() map[string]func(*http.Request) (*http.Response, error) // getTransportHealthy 函数返回一个健康状态的传输函数列表
}

// RoundTrip 是自定义负载均衡器的.RoundTrip方法，实现了http.RoundTripper接口。
// 它负责将HTTP请求发送到上游服务，并返回响应。
//
// 参数:
// req *http.Request - 需要发送的HTTP请求
//
// 返回值:
// *http.Response - 上游服务返回的HTTP响应
// error - 如果在发送请求过程中出现错误，则返回错误信息
func (c *customRoundTripperLoadBalancer) RoundTrip(req *http.Request) (*http.Response, error) {
	var roundTripper = c.getTransportHealthy()
	var upStreamServerSchemeAndHostOfName map[string]Pair[string, string] = map[string]Pair[string, string]{}
	for k := range roundTripper {
		upStreamServerSchemeAndHostOfName[k] = Pair[string, string]{c.upstreamURL.Scheme, c.upstreamURL.Host}
	}
	// 设置请求的Host为上游服务的Host
	req.Host = c.upstreamURL.Host

	PrintRequest(req) // 打印请求信息

	// 使用随机负载均衡策略选择一个健康状态的传输函数，并执行请求
	var rs, err = RandomLoadBalancer(roundTripper, req, upStreamServerSchemeAndHostOfName)

	if err != nil {
		fmt.Println("ERROR:", err) // 打印错误信息
	} else {
		PrintResponse(rs) // 打印响应信息
	}
	return rs, err
}

// randomShuffle 函数用于对指定的切片进行随机打乱。
// [T any] 表示该函数适用于任意类型的切片。
// arr []T 是输入的切片，函数会对其原地打乱顺序。
// 返回值 []T 是打乱后的切片。
func randomShuffle[T any](arr []T) []T {
	// 使用当前时间的纳秒级种子初始化随机数生成器，以确保每次运行结果都不同。
	rand.Seed(time.Now().UnixNano())
	// 使用 rand.Shuffle 函数来随机打乱切片的顺序。
	// 这个函数会传入切片的长度以及一个交换元素的函数。
	rand.Shuffle(len(arr), func(i, j int) {
		// 交换函数通过交换 arr[i] 和 arr[j] 来打乱顺序。
		arr[i], arr[j] = arr[j], arr[i]
	})
	return arr
}
func mapToArray[T comparable, Y any](m map[T]Y) []Pair[T, Y] {
	result := make([]Pair[T, Y], 0, len(m))
	for key := range m {
		result = append(result, Pair[T, Y]{First: key, Second: m[key]})
	}
	return result
}

// RandomLoadBalancer 是一个通过随机算法从提供的运输函数列表中选择一个来执行HTTP请求的负载均衡器。
// 参数：
// - roundTripper：一个包含多个http.RoundTripper函数的切片，这些函数将被用于发送HTTP请求。
// - req：指向待发送的HTTP请求的指针。
// 返回值：
// - *http.Response：从运输函数中返回的HTTP响应指针，如果所有运输函数都失败，则为nil。
// - error：如果在发送请求时遇到错误，则返回错误信息；否则为nil。
func RandomLoadBalancer(roundTripper map[string]func(*http.Request) (*http.Response, error), req *http.Request, upStreamServerSchemeAndHostOfName map[string]Pair[string, string]) (*http.Response, error) {
	// 打印传入的运输函数列表
	fmt.Println("RoundTripper:", roundTripper)

	PrintRequest(req)
	var roundTripperArray = mapToArray(roundTripper)
	// 使用随机算法对运输函数列表进行洗牌，以实现随机选择运输函数的效果
	var healthRoundTripper = randomShuffle(roundTripperArray)
	var rer error = nil

	// 遍历洗牌后的运输函数列表，尝试发送HTTP请求
	for _, transport := range healthRoundTripper {
		var name = transport.First
		var Scheme = upStreamServerSchemeAndHostOfName[name].First
		var Host = upStreamServerSchemeAndHostOfName[name].Second
		req.URL.Scheme = Scheme
		req.Host = Host
		req.URL.Host = Host
		var rs, err = transport.Second(req) // 执行运输函数
		if err != nil {
			// 如果请求发送失败，打印错误信息，并更新错误变量
			fmt.Println("ERROR:", err)
			rer = err

		} else {
			// 如果请求发送成功，打印响应信息，并返回响应和错误
			PrintResponse(rs)
			return rs, err
		}

	}
	// 如果所有运输函数尝试都失败，返回nil和错误信息
	return nil, rer
}

// PrintRequest 打印HTTP请求的详细信息
// 参数：
// req *http.Request - 代表一个HTTP请求的结构体指针
func PrintRequest(req *http.Request) {
	// 打印HTTP请求的基本信息
	fmt.Println(" HTTP Request {")
	fmt.Printf("Method: %s\n", req.Method) // 打印请求方法
	fmt.Printf("URL: %s\n", req.URL)       // 打印请求的URL
	fmt.Printf("Proto: %s\n", req.Proto)   // 打印请求的协议版本
	fmt.Printf("host: %s\n", req.Host)

	// 打印请求头信息
	fmt.Printf("Header: \n")
	PrintHeader(req.Header)

	fmt.Println("} HTTP Request ")
}

// PrintResponse 打印HTTP响应的详细信息
// 参数：
// resp *http.Response: 一个指向http.Response的指针，包含了HTTP响应的全部信息
func PrintResponse(resp *http.Response) {
	// 打印HTTP响应的起始标志
	fmt.Println(" HTTP Response {")
	// 打印响应的状态信息
	fmt.Printf("Status: %s\n", resp.Status)
	// 打印响应的状态码
	fmt.Printf("StatusCode: %d\n", resp.StatusCode)
	// 打印响应的协议版本
	fmt.Printf("Proto: %s\n", resp.Proto)
	// 打印响应的头部信息
	fmt.Printf("Header: \n")
	PrintHeader(resp.Header)

	// 打印HTTP响应的结束标志
	fmt.Println("} HTTP Response ")
}

// PrintHeader 打印HTTP头部信息
// 参数:
//
//	header http.Header - 要打印的HTTP头部信息
//
// 返回值:
//
//	无
func PrintHeader(header http.Header) {
	// 打印HTTP头部起始标签
	fmt.Println(" HTTP Header {")
	// 遍历头部信息，并打印每一条键值对
	for key, values := range header {
		fmt.Printf("%s: %v\n", key, values)
	}
	// 打印HTTP头部结束标签
	fmt.Println("} HTTP Header ")
}

type Pair[T any, Y any] struct {
	First  T
	Second Y
}
