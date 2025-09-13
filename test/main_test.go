// Package main - HTTP3/HTTP2 反向代理服务器测试
//
// 这个文件实现了一个完整的HTTP3/HTTP2反向代理服务器测试套件，主要功能包括：
//   - 支持HTTP/3和HTTP/2协议的反向代理
//   - 实现了基于单主机的负载均衡策略
//   - 提供主动和被动的健康检查机制
//   - 包含请求转发循环检测中间件
//   - 支持HTTPS和HTTP协议自动切换
//   - 集成了请求/响应日志打印功能
//   - 使用Gin框架作为HTTP服务器引擎
//
// 测试场景：
//   - 测试反向代理的基本功能
//   - 验证负载均衡策略
//   - 检测请求转发循环
//   - 测试健康检查机制
//   - 验证协议协商和头部处理
//
// 配置说明：
//   - 上游服务器：https://quic.nginx.org/
//   - HTTPS端口：18443
//   - 启用主动和被动健康检查
//
// 注意：此文件为测试文件，包含一些被注释的代码用于演示和测试目的。
package main

import (
	"bufio"
	"os"
	"testing"

	// "crypto/tls"
	"fmt"
	"log"

	// "math/rand"
	// "net"
	"net/http"
	"net/http/httptest"

	// "net/http/httputil"
	// "net/url"
	// "strconv"
	"strings"
	// "sync"
	// "time"

	"github.com/gin-gonic/gin"
	// "github.com/masx200/http3-reverse-proxy-server-experiment/adapter"
	// "github.com/masx200/http3-reverse-proxy-server-experiment/generic"
	"github.com/masx200/http3-reverse-proxy-server-experiment/load_balance"
	print_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/print"
	// "github.com/quic-go/quic-go"
	// "github.com/quic-go/quic-go/http3"
)

// 主程序入口
func TestMain(t *testing.T) {
	//健康检查过期时间毫秒
	// var maxAge = int64(5 * 1000)
	// 定义上游服务器地址
	/* 测试防环功能 */
	/*  remote error: tls: no application protocol */
	var upstreamServer = "https://quic.nginx.org/" //"https://assets.fastly.com/" //

	var LoadBalanceAndUpStream, err = load_balance.NewSingleHostHTTP3HTTP2LoadBalancerOfAddress(upstreamServer, upstreamServer)

	if err != nil {
		log.Fatal(err)
	}
	LoadBalanceAndUpStream.SetActiveHealthyCheckEnabled(true)
	LoadBalanceAndUpStream.SetPassiveHealthyCheckEnabled(true)
	var httpsPort = 18443
	// var httpPort = 18080
	// var upStreamServerSchemeAndHostOfName map[string]generic.PairInterface[string, string] = map[string]generic.PairInterface[string, string]{}
	engine := gin.Default()
	engine.Use(Forwarded(), LoopDetect())
	engine.Use(func(c *gin.Context) {

		c.Writer.Header().Add("Alt-Svc",
			"h3=\":"+fmt.Sprint(httpsPort)+"\";ma=86400,h3-29=\":"+fmt.Sprint(httpsPort)+"\";ma=86400,h3-27=\":"+fmt.Sprint(httpsPort)+"\";ma=86400",
		)
		c.Next()
	},
		func(ctx *gin.Context) {

			ctx.Writer.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			ctx.Next()
		})
	// // 定义上游服务器地址
	// var upstreamServers = []string{"https://quic.nginx.org/", "https://production.hello-word-worker.masx200.workers.dev/"}
	//打印上游
	// log.Println("Upstream servers:")
	// for _, server := range upstreamServers {
	// 	log.Println(server)
	// }

	// var expires = int64(0)
	// var upstreamServerOfName = map[string]string{}
	// var proxyServers map[string]func(*http.Request) (*http.Response, error) = map[string]func(*http.Request) (*http.Response, error){}

	// for _, urlString := range upstreamServers {
	// 	upstreamURL, err := url.Parse(urlString)
	// 	if err != nil {
	// 		log.Fatalf("Failed to parse upstream server URL: %v", err)
	// 	}
	// 	if upstreamURL.Path != "/" && upstreamURL.Path != "" {
	// 		log.Fatalf("upstreamServer Path must be / or empty")
	// 	}
	// 	proxy, err := createReverseProxy(urlString, int64(maxAge))

	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	proxyServers[urlString] = func(req *http.Request) (*http.Response, error) {
	// 		return proxy.Transport.RoundTrip(req)
	// 	}
	// 	upstreamServerOfName[urlString] = urlString
	// 	upStreamServerSchemeAndHostOfName[urlString] = generic.NewPairImplement[string, string](upstreamURL.Scheme, upstreamURL.Host)
	// }
	// 启动反向代理服务器
	// var healthyUpstream = proxyServers
	// var transportsUpstream = proxyServers
	// var mutex2 sync.Mutex
	// var getHealthyProxyServers = func() map[string]func(*http.Request) (*http.Response, error) {
	// 	return refreshHealthyUpStreams(func() int64 { return expires }, func() map[string]func(*http.Request) (*http.Response, error) { return healthyUpstream }, transportsUpstream, upstreamServerOfName, maxAge, func(i int64) { expires = i }, func(transportsUpstream map[string]func(*http.Request) (*http.Response, error)) {
	// 		healthyUpstream = transportsUpstream
	// 	}, &mutex2)

	// }
	engine.Any("/*path", func(c *gin.Context) {
		req := c.Request
		if req.TLS != nil {
			req.URL.Scheme = "https"
		} else {
			req.URL.Scheme = "http"
		}

		req.URL.Host = req.Host
		PrintRequest(req) // 打印请求信息

		// 使用随机负载均衡策略选择一个健康状态的传输函数，并执行请求
		var resp, err = LoadBalanceAndUpStream.RoundTrip(req)

		if err != nil {
			log.Println("ERROR:", err) // 打印错误信息
			c.AbortWithStatus(http.StatusBadGateway)
			return
		} else {
			PrintResponse(resp) // 打印响应信息
		}
		for k, vv := range resp.Header {
			for _, v := range vv {
				c.Writer.Header().Add(k, v)
			}
		}
		c.Status(resp.StatusCode)
		defer resp.Body.Close()
		bufio.NewReader(resp.Body).WriteTo(c.Writer)

	})
	engineMock := engine
	LoadBalanceAndUpStream.GetLoadBalanceService().Unwrap().HealthyCheckStart()
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		// 调用自定义处理器
		engineMock.Handler().ServeHTTP(w, req)
		var expectedStatus = 200
		// 验证响应状态码、响应头或响应体
		if w.Code != expectedStatus {
			t.Errorf("Expected status code %d, got %d", expectedStatus, w.Code)
		}
		PrintResponse(w.Result())

		w.Body.WriteTo(os.Stdout)
	}

	LoadBalanceAndUpStream.GetLoadBalanceService().Unwrap().HealthyCheckStop()
	// var hostname = "0.0.0.0"
	// server := &http.Server{
	// 	Addr: hostname + ":" + strconv.Itoa(httpsPort),
	// 	Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

	// 		engine.Handler().ServeHTTP(w, req) // 调用Gin引擎的Handler方法处理HTTP请求。

	// 	}), /*  &LoadBalanceHandler{
	// 		engine: engine,
	// 	}, */
	// }

	// go func() {
	// 	listener, err := net.Listen("tcp", hostname+":"+fmt.Sprint(httpPort))
	// 	if err != nil {
	// 		log.Fatal("ListenAndServe: ", err)
	// 	}
	// 	log.Printf("http reverse proxy server started on port %s", listener.Addr())

	// 	// 设置自定义处理器
	// 	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
	// 		engine.Handler().ServeHTTP(w, req)
	// 	})

	// 	// 开始服务
	// 	err = http.Serve(listener, nil)
	// 	if err != nil {
	// 		log.Fatal("Serve: ", err)
	// 	}
	// }()
	// certFile := "cert.crt"
	// keyFile := "key.pem"
	// go func() {
	// 	var handlerFunc = func(w http.ResponseWriter, req *http.Request) {
	// 		engine.Handler().ServeHTTP(w, req)
	// 	}

	// 	bCap := hostname + ":" + fmt.Sprint(httpsPort)
	// 	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	// 		handlerFunc(w, req)
	// 	})
	// 	server := http3.Server{
	// 		Handler:    handler,
	// 		Addr:       bCap,
	// 		QuicConfig: &quic.Config{
	// 			// Tracer: qlog.DefaultTracer,
	// 		},
	// 	}
	// 	log.Printf("Starting http3 reverse proxy server on " + hostname + ":" + strconv.Itoa(httpsPort))

	// 	var err = server.ListenAndServeTLS(certFile, keyFile)
	// 	// var err = http3.ListenAndServe(bCap, certFile, keyFile, &
	// 	// 	serveHTTP: handlerFunc,
	// 	// })
	// 	if err != nil {
	// 		log.Fatal("Serve: ", err)
	// 	}
	// }()
	// log.Printf("Starting https reverse proxy server on " + hostname + ":" + strconv.Itoa(httpsPort))

	// errx := server.ListenAndServeTLS(certFile, keyFile)
	// if errx != nil {
	// 	log.Fatal(errx)
	// }
	// // log.Fatal(errx)
}

// Forwarded 创建并返回一个 gin.HandlerFunc，用于在 HTTP 请求的 Header 中添加 "Forwarded" 信息。
// 这个信息包含了客户端的 IP 地址、代理的标识、原始请求的目标主机名以及使用的协议（HTTP 或 HTTPS）。
// 返回值是一个处理 gin.Context 的函数。
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

// ForwardedBy 结构体
//
// 描述：这个结构体用于表示一个转发标识。
//
// 字段：
// Identifier string - 用于唯一标识转发的字符串。
type ForwardedBy struct {
	Identifier string
}

// 辅助函数：将ForwardedBy列表转换为集合（set），用于快速判断重复项
func setFromForwardedBy(forwardedByList []ForwardedBy) map[string]bool {
	set := make(map[string]bool)
	for _, fb := range forwardedByList {
		set[fb.Identifier] = true
	}
	return set
}

// parseForwardedHeader 解析 "Forwarded" HTTP 头部信息，返回一个 ForwardedBy 结构体切片。
// header: 代表被转发的请求的 "Forwarded" 头部字符串。
// 返回值: 一个包含所有转发标识的 ForwardedBy 结构体切片，以及可能发生的错误。

func parseForwardedHeader(header string) ([]ForwardedBy, error) {
	var forwardedByList []ForwardedBy
	parts := strings.Split(header, ", ")

	for _, part := range parts {
		for _, param := range strings.Split(part, ";") {
			param = strings.TrimSpace(param)
			if !strings.HasPrefix(param, "by=") {
				continue
			}

			// 分离 by 参数的值
			value := strings.TrimPrefix(param, "by=")
			// host, port, err := net.SplitHostPort(value)
			// if err != nil {
			// 如果没有端口信息，host 就是整个值
			var host = value
			// port = ""
			// }

			forwardedBy := ForwardedBy{
				Identifier: host,
				// Port:       port,
			}

			// 检查是否重复
			// isDuplicate := false
			// for _, existing := range forwardedByList {
			// 	if existing.Identifier == forwardedBy.Identifier && existing.Port == forwardedBy.Port {
			// 		isDuplicate = true
			// 		break
			// 	}
			// }
			// if !isDuplicate {
			forwardedByList = append(forwardedByList, forwardedBy)
			// }
		}
	}

	return forwardedByList, nil
}

// LoopDetect 是一个用于检测请求中'Forwarded'头是否存在重复'by'标识符的gin中间件。
// 如果发现重复的'by'标识符，将返回状态码508并提供错误信息。
// 如果无法解析'Forwarded'头，将返回状态码400并给出解析错误的具体信息。
// 返回值为一个gin.HandlerFunc类型的函数，可直接用于gin路由的中间件配置。

func LoopDetect() gin.HandlerFunc {
	return func(c *gin.Context) {
		var r = c.Request
		var w = c.Writer
		forwardedHeader := strings.Join(r.Header.Values("Forwarded"), ", ")
		log.Println("forwardedHeader:", forwardedHeader)
		forwardedByList, err := parseForwardedHeader(forwardedHeader)
		log.Println("forwardedByList:", forwardedByList)
		if len(forwardedByList) != len(setFromForwardedBy(forwardedByList)) {
			w.WriteHeader(508)
			fmt.Fprintln(w, "Duplicate 'by' identifiers found in 'Forwarded' header.")
			log.Println("Duplicate 'by' identifiers found in 'Forwarded' header.")
			c.Abort()
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error parsing 'Forwarded' header: %v", err)
			return
		}
		c.Next()
	}
}

func PrintRequest(req *http.Request) {
	print_experiment.PrintRequest(req)
}

// PrintRequest 打印HTTP请求的详细信息
// 参数：
// req *http.Request - 代表一个HTTP请求的结构体指针

func PrintHeader(header http.Header) {
	print_experiment.PrintHeader(header)
}

// Pair是一个泛型结构体，用于存储一对任意类型的值。
// T和Y是泛型参数，代表First和Second可以是任何类型。

func PrintResponse(resp *http.Response) {
	print_experiment.PrintResponse(resp)
}
