package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"runtime"
	"sync"
	"time"

	// "crypto/tls"
	"fmt"
	"log"

	// "math/rand"
	"net"
	"net/http"
	"net/url"

	// "net/http/httputil"
	// "net/url"
	"strconv"
	"strings"

	// "sync"
	// "time"
	pprofhttp "net/http/pprof"
	"runtime/pprof"

	"github.com/gin-gonic/gin"
	"github.com/moznion/go-optional"

	// "github.com/masx200/http3-reverse-proxy-server-experiment/adapter"
	// "github.com/masx200/http3-reverse-proxy-server-experiment/generic"
	"github.com/masx200/http3-reverse-proxy-server-experiment/adapter"
	h3_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/h3"
	"github.com/masx200/http3-reverse-proxy-server-experiment/http2_only"
	print_experiment "github.com/masx200/http3-reverse-proxy-server-experiment/print"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func CreateHTTP2CRoundTripperOfUpStreamServer() http.RoundTripper {

	return &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}
}

type HTTPRoundTripperMiddleWare = func(req *http.Request, next func(req *http.Request) (*http.Response, error)) (*http.Response, error)

// CreateHTTPRoundTripperMiddleWareOfUpStreamServerURL 创建一个用于修改HTTP请求的中间件，使其指向指定的上游服务器URL。
//
// 参数:
//
//	upstreamServerURLstring string - 上游服务器的URL字符串。
//
// 返回值:
//
//	HTTPRoundTripperMiddleWare - 一个函数，接受一个http.Request和一个next函数作为参数，返回一个(*http.Response, error)。
func CreateHTTPRoundTripperMiddleWareOfUpStreamServerURL(upstreamServerURLstring string) HTTPRoundTripperMiddleWare {
	return func(req *http.Request, next func(req *http.Request) (*http.Response, error)) (*http.Response, error) {
		//parse url of upstreamServerURL
		upstreamServerURL, err := url.Parse(upstreamServerURLstring)
		if err != nil {
			return nil, err
		}

		req.Host = upstreamServerURL.Host
		req.URL.Scheme = upstreamServerURL.Scheme
		req.URL.Host = upstreamServerURL.Host
		return next(req)
	}
}

// 主程序入口
func main() {
	strArgupstreamServer := flag.String("upstream-server", "", "upstream-server,example \"https://workers.cloudflare.com/\"")
	intArghttpPort := flag.Int("http-port", 18080, "http-port")
	int2ArghttpsPort := flag.Int("https-port", 18443, "https-port")
	StringArgprotocol := flag.String("upstream-protocol", "h3", "upstream-protocol,supports (h3,h2,h2c,http/1.1)")
	tlscertArg := flag.String("tls-cert", "cert.crt", "tls-cert")
	tlskeyArg := flag.String("tls-key", "key.pem", "tls-key")
	Arglistenhostname := flag.String("listen-hostname", "0.0.0.0", "listen-hostname")
	tlsboolArg := flag.Bool("listen-tls", true, "listen-tls")
	Arglistenhttp := flag.Bool("listen-http", true, "listen-http")
	Arglistenh2c := flag.Bool("listen-h2c", true, "listen-h2c")
	Arglistenhttp3 := flag.Bool("listen-http3", true, "listen-http3")
	Arg_debug_pprof := flag.Bool("debug-pprof", false, "debug-pprof")
	// 解析命令行参数
	flag.Parse()

	// 输出解析后的参数值
	log.Printf("debug-pprof argument: %v\n", *Arg_debug_pprof)
	log.Printf("listen-hostname argument: %v\n", *Arglistenhostname)
	log.Printf("listen-http argument: %v\n", *Arglistenhttp)
	log.Printf("listen-h2c argument: %v\n", *Arglistenh2c)
	log.Printf("listen-http3 argument: %v\n", *Arglistenhttp3)
	log.Printf("tls-cert argument: %s\n", *tlscertArg)
	log.Printf("tls-key argument: %s\n", *tlskeyArg)
	log.Printf("upstream-server argument: %s\n", *strArgupstreamServer)
	log.Printf("http-port argument: %d\n", *intArghttpPort)
	log.Printf("https-port argument: %d\n", *int2ArghttpsPort)
	log.Printf("upstream-protocol argument: %v\n", *StringArgprotocol)
	log.Printf("listen-tls argument: %v\n", *tlsboolArg)
	var upstreamServer = *strArgupstreamServer
	if len(upstreamServer) == 0 {
		log.Fatal("error :upstream-server is empty")
	}
	var upstreamServerDefaultTransport http.RoundTripper

	if strings.Contains(*StringArgprotocol, "h3") {
		upstreamServerDefaultTransport = CreateHTTP3RoundTripperOfUpStreamServer(upstreamServer)
	} else if strings.Contains(*StringArgprotocol, "h2c") {
		var rt = CreateHTTP2CRoundTripperOfUpStreamServer()
		upstreamServerDefaultTransport = adapter.RoundTripTransport(func(r *http.Request) (*http.Response, error) {
			return CreateHTTPRoundTripperMiddleWareOfUpStreamServerURL(upstreamServer)(r, rt.RoundTrip)
		})
	} else if !strings.Contains(*StringArgprotocol, "h3") {
		var rt = CreateHTTP12RoundTripperOfUpStreamServer(strings.Split(*StringArgprotocol, ","))
		upstreamServerDefaultTransport = adapter.RoundTripTransport(func(r *http.Request) (*http.Response, error) {
			return CreateHTTPRoundTripperMiddleWareOfUpStreamServerURL(upstreamServer)(r, rt.RoundTrip)
		})
	}
	//健康检查过期时间毫秒
	// var maxAge = int64(5 * 1000)
	// 定义上游服务器地址
	/* 测试防环功能 */
	// var upstreamServers = []string{ /* "https://production.hello-word-worker-cloudflare.masx200.workers.dev/", "https://fastly-compute-hello-world-javascript.edgecompute.app/" */ }
	var httpsPort = *int2ArghttpsPort
	var httpPort = *intArghttpPort
	// var upStreamServerSchemeAndHostOfName map[string]generic.PairInterface[string, string] = map[string]generic.PairInterface[string, string]{}
	engine := gin.Default()
	engine.Use(Forwarded(), LoopDetect())
	engine.Use(func(c *gin.Context) {
		if *Arglistenhttp3 {
			c.Writer.Header().Add("Alt-Svc",
				"h3=\":"+fmt.Sprint(httpsPort)+"\";ma=886400,h3-29=\":"+fmt.Sprint(httpsPort)+"\";ma=886400,h3-27=\":"+fmt.Sprint(httpsPort)+"\";ma=886400",
			)
		}
		if *Arglistenh2c {
			c.Writer.Header().Add("Alt-Svc",
				"h2c=\":"+fmt.Sprint(httpPort)+"\";ma=886400",
			)
		}

		c.Next()
	},
		func(ctx *gin.Context) {
			if *tlsboolArg {
				ctx.Writer.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			ctx.Next()
		})
	var debug_pprof_app = optional.None[*gin.Engine]()
	if *Arg_debug_pprof {

		debug_pprof_app = CreateDebugPprofApplication()

	}

	engine.Use(func(ctx *gin.Context) {
		if *Arg_debug_pprof {
			if strings.HasPrefix(ctx.Request.URL.Path, "/debug/pprof/") && debug_pprof_app.IsSome() {
				debug_pprof_app.Unwrap().ServeHTTP(ctx.Writer, ctx.Request)
				ctx.Abort()
				return

			}
		}
		req := ctx.Request
		if req.TLS != nil {
			req.URL.Scheme = "https"
		} else {
			req.URL.Scheme = "http"
		}

		req.URL.Host = req.Host
		PrintRequest(req) // 打印请求信息

		// 使用随机负载均衡策略选择一个健康状态的传输函数，并执行请求

		var resp, err = upstreamServerDefaultTransport.RoundTrip(req) //RandomLoadBalancer(getHealthyProxyServers(), req, upStreamServerSchemeAndHostOfName)

		if err != nil {
			log.Println("ERROR:", err) // 打印错误信息
			ctx.String(502, "ERROR: "+err.Error())
			ctx.AbortWithStatus(http.StatusBadGateway)
			return
		} else {
			PrintResponse(resp) // 打印响应信息
		}
		for k, vv := range resp.Header {
			for _, v := range vv {
				ctx.Writer.Header().Add(k, v)
			}
		}
		ctx.Status(resp.StatusCode)
		defer resp.Body.Close()
		bufio.NewReader(resp.Body).WriteTo(ctx.Writer)
		ctx.Abort()

	})
	var hostname = *Arglistenhostname //"0.0.0.0"

	// go func() {

	// 	if *tlsboolArg && *Arglistenhttp {

	// 		listener, err := net.Listen("tcp", hostname+":"+fmt.Sprint(httpPort))
	// 		if err != nil {
	// 			log.Fatal("ListenAndServe: ", err)
	// 		}
	// 		log.Printf("http reverse proxy server started on port %s", listener.Addr())

	// 		// 设置自定义处理器
	// 		var handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	// 			engine.Handler().ServeHTTP(w, req)
	// 		})
	// 		h2s := &http2.Server{
	// 			// ...
	// 		}
	// 		// 开始服务
	// 		err = http.Serve(listener, h2c.NewHandler(handler, h2s))
	// 		if err != nil {
	// 			log.Fatal("Serve: ", err)
	// 		}
	// 	}

	// }()

	var group sync.WaitGroup
	certFile := *tlscertArg //"cert.crt"
	keyFile := *tlskeyArg   // "key.pem"
	group.Add(1)
	go func() {

		defer group.Done()
		if *Arglistenhttp3 {
			var handlerFunc = func(w http.ResponseWriter, req *http.Request) {
				engine.Handler().ServeHTTP(w, req)
			}

			bCap := hostname + ":" + fmt.Sprint(httpsPort)
			handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				handlerFunc(w, req)
			})
			server := http3.Server{
				Handler:    handler,
				Addr:       bCap,
				QUICConfig: &quic.Config{
					// Tracer: qlog.DefaultTracer,
				},
			}
			log.Printf("Starting http3 reverse proxy server on " + hostname + ":" + strconv.Itoa(httpsPort))

			var err = server.ListenAndServeTLS(certFile, keyFile)
			// var err = http3.ListenAndServe(bCap, certFile, keyFile, &
			// 	serveHTTP: handlerFunc,
			// })
			if err != nil {
				log.Fatal("Serve: ", err)
			}
		}

	}()
	group.Add(1)
	go func() {
		defer group.Done()
		if *tlsboolArg {
			log.Printf("Starting https reverse proxy server on " + hostname + ":" + strconv.Itoa(httpsPort))
			server := &http.Server{
				Addr: hostname + ":" + strconv.Itoa(httpsPort),
				Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					engine.Handler().ServeHTTP(w, req) // 调用Gin引擎的Handler方法处理HTTP请求。

				}), /*  &LoadBalanceHandler{
					engine: engine,
				}, */
			}
			errx := server.ListenAndServeTLS(certFile, keyFile)
			if errx != nil {
				log.Fatal(errx)
			}
		}
	}()
	group.Add(1)
	go func() {
		defer group.Done()
		if *Arglistenhttp || *Arglistenh2c {

			listener, err := net.Listen("tcp", hostname+":"+fmt.Sprint(httpPort))
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
			log.Printf("http reverse proxy server started on port %s", listener.Addr())

			// 设置自定义处理器
			var handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				engine.Handler().ServeHTTP(w, req)
			})
			http2Server := &http2.Server{
				// ...
			}

			if *Arglistenh2c {
				if *Arglistenhttp {
					err = http.Serve(listener, h2c.NewHandler(handler, http2Server))
				} else {
					err = http.Serve(listener, http2_only.NewHandler(handler, http2Server))
				}
			} else {
				err = http.Serve(listener, (handler))
			}
			// 开始服务

			if err != nil {
				log.Fatal("Serve: ", err)
			}
		}
	}()
	group.Wait()
}

// CreateDebugPprofApplication 创建并配置一个启用PPROF的Gin路由器实例。
// 该函数不接受参数，返回一个包含gin.Engine指针的optional.Option类型。
// 这允许调用者选择是否使用PPROF调试功能。
func CreateDebugPprofApplication() optional.Option[*gin.Engine] {
	var debug_pprof_app = (gin.Default())
	debug_pprof_app.Any("/debug/pprof/allocs", func(ctx *gin.Context) {
		pprofhttp.Handler("allocs").ServeHTTP(ctx.Writer, ctx.Request)
		ctx.Abort()
	})
	debug_pprof_app.Any("/debug/pprof/mutex", func(ctx *gin.Context) {
		pprofhttp.Handler("mutex").ServeHTTP(ctx.Writer, ctx.Request)
		ctx.Abort()
	})
	debug_pprof_app.Any("/debug/pprof/threadcreate", func(ctx *gin.Context) {
		pprofhttp.Handler("threadcreate").ServeHTTP(ctx.Writer, ctx.Request)
		ctx.Abort()
	})
	debug_pprof_app.Any("/debug/pprof/", func(ctx *gin.Context) {
		pprofhttp.Index(ctx.Writer, ctx.Request)
		ctx.Abort()
	})
	debug_pprof_app.Any("/debug/pprof/block", func(ctx *gin.Context) {
		pprofhttp.Handler("block").ServeHTTP(ctx.Writer, ctx.Request)
		ctx.Abort()
	})
	debug_pprof_app.Any("/debug/pprof/goroutine", func(ctx *gin.Context) {
		var f = ctx.Writer

		runtime.GC()
		if err := pprof.Lookup("goroutine").WriteTo(f, 1); err != nil {
			log.Println("could not start goroutine profile: ", err)
			f.Write([]byte("could not start goroutine profile: " + err.Error()))
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Abort()
	})
	debug_pprof_app.Any("/debug/pprof/cmdline", func(ctx *gin.Context) {
		pprofhttp.Cmdline(ctx.Writer, ctx.Request)
		ctx.Abort()
	})
	debug_pprof_app.Any("/debug/pprof/profile", func(ctx *gin.Context) {
		pprofhttp.Profile(ctx.Writer, ctx.Request)
		ctx.Abort()
	})
	debug_pprof_app.Any("/debug/pprof/symbol", func(ctx *gin.Context) {
		pprofhttp.Symbol(ctx.Writer, ctx.Request)
		ctx.Abort()
	})
	debug_pprof_app.Any("/debug/pprof/trace", func(ctx *gin.Context) {
		pprofhttp.Trace(ctx.Writer, ctx.Request)
		ctx.Abort()
	})
	debug_pprof_app.Any("/debug/pprof/heap", func(ctx *gin.Context) {
		var f = ctx.Writer

		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Println("could not start heap profile: ", err)
			f.Write([]byte("could not start heap profile: " + err.Error()))
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Abort()
	})
	return optional.Some(debug_pprof_app)
}

func CreateHTTP12RoundTripperOfUpStreamServer(alpns []string) http.RoundTripper {
	if len(alpns) > 0 {
		return &http.Transport{TLSClientConfig: &tls.Config{
			// ServerName: ,
			NextProtos: alpns}, ForceAttemptHTTP2: true}
	}
	return &http.Transport{ForceAttemptHTTP2: true}
}

// CreateHTTP3RoundTripperOfUpStreamServer 创建一个指向上游服务器的HTTP/3轮询器。
// 参数:
//
//	upstreamServer string - 上游服务器的URL。
//
// 返回值:
//
//	adapter.HTTPRoundTripperAndCloserInterface - 支持HTTP/3协议的轮询器接口，可用于发起HTTP请求。
func CreateHTTP3RoundTripperOfUpStreamServer(upstreamServer string) adapter.HTTPRoundTripperAndCloserInterface {
	var mutex sync.Mutex
	var started = false
	var h3rt = h3_experiment.CreateHTTP3TransportWithIPGetter(func() (string, error) {
		upstreamServerURL, err := url.Parse(upstreamServer)
		if err != nil {
			return "", err
		}
		return upstreamServerURL.Hostname(), nil
	})
	log.Println(
		"INFO: Creating new HTTP/3 round tripper for upstream server",
	)
	var oldH3rt optional.Option[adapter.HTTPRoundTripperAndCloserInterface] = nil

	ticker := time.NewTicker(time.Minute)
	var intervalTaskStart = func() {
		mutex.Lock()
		defer mutex.Unlock()
		if started {
			return
		}
		started = true
		go func() {
			for range ticker.C {
				log.Println(
					"INFO: Creating new HTTP/3 round tripper for upstream server",
				)
				oldH3rt = optional.Some(h3rt)
				h3rt = h3_experiment.CreateHTTP3TransportWithIPGetter(func() (string, error) {
					upstreamServerURL, err := url.Parse(upstreamServer)
					if err != nil {
						return "", err
					}
					return upstreamServerURL.Hostname(), nil
				})

				if oldH3rt != nil && oldH3rt.IsSome() {
					oldH3rt.Unwrap().Close()
					log.Println(
						"INFO: Closed old HTTP/3 round tripper for upstream server",
					)
				}
			}
		}()

	}
	//为了防止udp被限速,需要定时更换端口

	// 每分钟更换一个新的 http3.RoundTripper 并关闭旧的

	upstreamServerDefaultTransport := &adapter.HTTPRoundTripperAndCloserImplement{RoundTripper: (func(req *http.Request) (*http.Response, error) {
		return CreateHTTPRoundTripperMiddleWareOfUpStreamServerURL(upstreamServer)(req, func(req *http.Request) (*http.Response, error) {
			if !started {
				go intervalTaskStart()
			}

			return h3rt.RoundTrip(req)
		})

	}), Closer: func() error {
		ticker.Stop()
		if oldH3rt != nil && oldH3rt.IsSome() {
			oldH3rt.Unwrap().Close()
		}

		return h3rt.Close()
	}}
	return upstreamServerDefaultTransport
}

// checkUpstreamHealth 检查上游服务的健康状态
// url: 要检查的服务的URL地址
// RoundTrip: 用于发送HTTP请求的函数，模拟HTTP客户端的行为
// 返回值: 返回一个布尔值，表示上游服务的健康状态，true为健康，false为不健康

// func checkUpstreamHealth(url string, RoundTrip func(*http.Request) (*http.Response, error)) (bool, error) {

// 	// 发送HEAD请求并检查返回的状态码
// 	statusCode, err := sendHeadRequestAndCheckStatus(url, RoundTrip)
// 	if err != nil {
// 		log.Printf("Error: %v\n", err)
// 		return false, err
// 	}

// 	// 打印状态码信息
// 	log.Printf("health check Status code: %d\n", statusCode)

// 	// 根据状态码判断上游服务的健康状态
// 	if statusCode < 500 {
// 		log.Println("Status code is less than 500.")
// 		return true, nil
// 	} else {
// 		log.Println("ERROR:"+"Status code is 500 or greater.", statusCode)

// 	}
// 	return false,  errors.New("ERROR:Status code is 500 or greater" + fmt.Sprint(statusCode))
// }

// sendHeadRequestAndCheckStatus 发送一个HEAD请求并检查状态码。
// url: 请求的目标URL。
// RoundTrip: 自定义的HTTP.RoundTripper函数，用于发送请求。
// 返回值: 请求的状态码和可能出现的错误。
// func sendHeadRequestAndCheckStatus(url string, RoundTrip func(*http.Request) (*http.Response, error)) (int, error) {
// 	client := &http.Client{}

// 	client.Transport = adapter.RoundTripTransport(RoundTrip)
// 	req, err := http.NewRequest("HEAD", url, nil)
// 	if err != nil {
// 		return 0, err
// 	}
// 	PrintRequest(req)
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer resp.Body.Close()
// 	PrintResponse(resp)
// 	return resp.StatusCode, nil
// }

// refreshHealthyUpStreams加锁操作
// var mutex sync.Mutex

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

// LoadBalanceHandler 是一个用于负载均衡处理的结构体。

// ServeHTTP 是一个实现http.Handler接口的方法，用于处理HTTP请求。
// 参数w是用于向客户端发送响应的http.ResponseWriter，
// 参数req是客户端发来的HTTP请求。
// ChangeURLScheme 用于更改给定URL的协议。
//
// 参数：
// originalURLStr: 原始URL字符串，需要被更改协议的URL。
// newScheme: 新的协议名称，例如："https"。
//
// 返回值：
// 返回更改协议后的URL字符串和一个error。
// 如果解析原始URL或生成新URL时发生错误，将返回一个非空的error。

// func ChangeURLScheme(originalURLStr string, newScheme string) (string, error) {
// 	// Parse the original URL
// 	originalURL, err := url.Parse(originalURLStr)
// 	if err != nil {
// 		return "",  errors.New("failed to parse the original URL: %v", err)
// 	}

// 	// Modify the scheme
// 	originalURL.Scheme = newScheme

// 	// Return the string representation of the modified URL
// 	return originalURL.String(), nil
// }

// createReverseProxy 创建一个反向代理，根据上游服务器的健康状况动态选择HTTP/3或HTTP/2进行通信。
//
// 参数:
// upstreamServer - 上游服务器的URL字符串。
//
// 返回值:
// *httputil.ReverseProxy - 配置好的反向代理实例。
// error - 创建过程中遇到的任何错误。
// func CreateReverseProxy(upstreamServer string, protocol string) (*httputil.ReverseProxy, error) {
// 	// 解析上游服务器URL，确保其路径为根路径或为空
// 	upstreamURL, err := url.Parse(upstreamServer)
// 	if err != nil {
// 		log.Printf("Failed to parse upstream server URL: %v", err)
// 		return nil, err
// 	}
// 	if upstreamURL.Path != "/" && upstreamURL.Path != "" {
// 		log.Printf("upstreamServer Path must be / or empty")
// 		return nil, err
// 	}

// 	// 初始化HTTP/3客户端
// 	http3Client := &http.Client{
// 		Transport: &http3.RoundTripper{
// 			TLSClientConfig: &tls.Config{},
// 			QuicConfig:      &quic.Config{},
// 		},
// 	}
// 	// 初始化HTTP/2客户端，使用默认传输
// 	http2Client := &http.Client{
// 		Transport: http.DefaultTransport,
// 	}
// 	// http3url, err := ChangeURLScheme(upstreamServer, "http3")
// 	// if err != nil {
// 	// 	return nil,  errors.New("failed to parse upstream server URL: %v", err)
// 	// }
// 	// var http12url string
// 	// if upstreamURL.Scheme == "http" {
// 	// 	http12urlll, err := ChangeURLScheme(upstreamServer, "http1")
// 	// 	if err != nil {
// 	// 		return nil,  errors.New("failed to parse upstream server URL: %v", err)
// 	// 	}
// 	// 	http12url = http12urlll
// 	// } else {
// 	// 	http12urlll, err := ChangeURLScheme(upstreamServer, "http2")
// 	// 	if err != nil {
// 	// 		return nil,  errors.New("failed to parse upstream server URL: %v", err)
// 	// 	}
// 	// 	http12url = http12urlll
// 	// }

// 	// 定义上游服务器的传输方式，包括HTTP/3和HTTP/2
// 	// var transportsUpstream = map[string]func(*http.Request) (*http.Response, error){
// 	// 	http3url: func(req *http.Request) (*http.Response, error) { return http3Client.Transport.RoundTrip(req) },

// 	// 	http12url: func(req *http.Request) (*http.Response, error) { return http2Client.Transport.RoundTrip(req) },
// 	// }
// 	// var upstreamServerOfName = map[string]string{}

// 	// for k := range transportsUpstream {
// 	// 	upstreamServerOfName[k] = upstreamServer
// 	// }
// 	// 设置健康检查的超时时间毫秒

// 	// var expires = int64(0)
// 	// var healthyUpstream = transportsUpstream
// 	// var mutex2 sync.Mutex
// 	// 自定义负载均衡的传输器，根据上游服务器的健康状况选择传输方式
// 	customTransport := &customRoundTripperLoadBalancer{
// 		// upstreamURL: upstreamURL,
// 		// getTransportHealthy: func() map[string]func(*http.Request) (*http.Response, error) {
// 		// 	// 对上游服务器进行健康检查，选择健康的传输方式
// 		// 	return refreshHealthyUpStreams(func() int64 { return expires }, func() map[string]func(*http.Request) (*http.Response, error) { return healthyUpstream }, transportsUpstream, upstreamServerOfName, maxAge, func(i int64) { expires = i }, func(transportsUpstream map[string]func(*http.Request) (*http.Response, error)) {
// 		// 		healthyUpstream = transportsUpstream
// 		// 	}, &mutex2)
// 		// },
// 	}

// 	// 初始化反向代理
// 	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)

// 	// 设置反向代理的传输器为自定义的负载均衡传输器
// 	proxy.Transport = customTransport
// 	return proxy, err
// }

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
// func refreshHealthyUpStreams(getExpires func() int64, getHealthyUpstream func() map[string]func(*http.Request) (*http.Response, error), transportsUpstream map[string]func(*http.Request) (*http.Response, error), upstreamServerOfName map[string]string, maxAge int64, setExpires func(int64), setHealthyUpstream func(transportsUpstream map[string]func(*http.Request) (*http.Response, error)), mutex2 *sync.Mutex) map[string]func(*http.Request) (*http.Response, error) {
// 	// mutex.Lock()
// 	// defer mutex.Unlock()

// 	// 检查当前上游服务器列表是否已过期。
// 	if getExpires() > time.Now().UnixMilli() {
// 		log.Println("不需要进行健康检查", "还剩余的时间毫秒", getExpires()-time.Now().UnixMilli())
// 		log.Println("健康的上游服务器", getHealthyUpstream())
// 		return getHealthyUpstream()
// 	}

// 	// 在后台进行健康检查更新。
// 	go func() {

// 		mutex2.Lock()
// 		defer mutex2.Unlock()
// 		if getExpires() > time.Now().UnixMilli() {
// 			log.Println("不需要进行健康检查", "还剩余的时间毫秒", getExpires()-time.Now().UnixMilli())
// 			log.Println("健康的上游服务器", getHealthyUpstream())
// 			return
// 		}
// 		var healthy = map[string]func(*http.Request) (*http.Response, error){}
// 		log.Println("需要进行健康检查", "已经过期的时间毫秒", -getExpires()+time.Now().UnixMilli())
// 		//需要并行检查
// 		// 遍历所有上游服务器进行健康检查。
// 		var promises = make(chan struct{}, len(transportsUpstream))
// 		for key, roundTrip := range transportsUpstream {
// 			keyi0 := key
// 			roundTripi0 := roundTrip

// 			go func() {
// 				defer func() {
// 					promises <- struct{}{}
// 				}()
// 				var upstreamServer = upstreamServerOfName[keyi0]
// 				//loop variable roundTrip captured by func literal loop closure
// 				if ok, err := checkUpstreamHealth(upstreamServer, roundTripi0); ok {
// 					healthy[keyi0] = roundTripi0
// 					log.Println("健康检查成功", keyi0, upstreamServer)
// 				} else {

// 					log.Println("健康检查失败", keyi0, upstreamServer, err)
// 				}

// 			}()

// 		}
// 		for range transportsUpstream {
// 			<-promises
// 		}
// 		// 根据健康检查结果更新健康上游服务器列表。
// 		if len(healthy) == 0 {
// 			setHealthyUpstream(transportsUpstream)
// 			log.Println("没有健康的上游服务器", getHealthyUpstream())
// 		} else {
// 			setHealthyUpstream(healthy)
// 			log.Println("找到健康的上游服务器", getHealthyUpstream())
// 		}

// 		// 设置上游服务器列表的新过期时间。
// 		var expires = time.Now().UnixMilli() + int64(maxAge)
// 		setExpires(expires)
// 	}()

// 	// 返回当前的健康上游服务器列表。
// 	return getHealthyUpstream()
// }

// customRoundTripperLoadBalancer 是一个自定义的负载均衡器，
// 用于在多个上游服务之间进行请求的负载均衡。
// type customRoundTripperLoadBalancer struct {
// 	upstreamURL         *url.URL                                                      // upstreamURL 指定了上游服务的URL
// 	getTransportHealthy func() map[string]func(*http.Request) (*http.Response, error) // getTransportHealthy 函数返回一个健康状态的传输函数列表
// }

// RoundTrip 是自定义负载均衡器的.RoundTrip方法，实现了http.RoundTripper接口。
// 它负责将HTTP请求发送到上游服务，并返回响应。
//
// 参数:
// req *http.Request - 需要发送的HTTP请求
//
// 返回值:
// *http.Response - 上游服务返回的HTTP响应
// error - 如果在发送请求过程中出现错误，则返回错误信息
// func (c *customRoundTripperLoadBalancer) RoundTrip(req *http.Request) (*http.Response, error) {
// 	var roundTripper = c.getTransportHealthy()
// 	var upStreamServerSchemeAndHostOfName map[string]generic.PairInterface[string, string] = map[string]generic.PairInterface[string, string]{}
// 	for k := range roundTripper {
// 		upStreamServerSchemeAndHostOfName[k] = generic.NewPairImplement[string, string](c.upstreamURL.Scheme, c.upstreamURL.Host)
// 	}
// 	// 设置请求的Host为上游服务的Host
// 	req.Host = c.upstreamURL.Host

// 	PrintRequest(req) // 打印请求信息

// 	// 使用随机负载均衡策略选择一个健康状态的传输函数，并执行请求
// 	var rs, err = RandomLoadBalancer(roundTripper, req, upStreamServerSchemeAndHostOfName)

// 	if err != nil {
// 		log.Println("ERROR:", err) // 打印错误信息
// 	} else {
// 		PrintResponse(rs) // 打印响应信息
// 	}
// 	return rs, err
// }

// randomShuffle 函数用于对指定的切片进行随机打乱。
// [T any] 表示该函数适用于任意类型的切片。
// arr []T 是输入的切片，函数会对其原地打乱顺序。
// 返回值 []T 是打乱后的切片。
// func randomShuffle[T any](arr []T) []T {
// 	// 使用当前时间的纳秒级种子初始化随机数生成器，以确保每次运行结果都不同。
// 	return generic.RandomShuffle(arr)
// } // mapToArray 将一个映射（map）转换为包含键值对（Pair）的切片（slice）。
// 参数 m 是一个类型为 map[T]Y 的映射，其中 T 是可比较的类型，Y 是任意类型。
// 返回值是一个类型为 []Pair[T, Y] 的切片，其中 Pair 是一个包含两个字段 First 和 Second 的结构体。
// 这个函数主要用于将映射的键值对形式转换为切片形式，方便后续处理。

// func mapToArray[T comparable, Y any](m map[T]Y) []generic.PairInterface[T, Y] {
// 	result := make([]generic.PairInterface[T, Y], 0, len(m))
// 	for key := range m {
// 		result = append(result, generic.NewPairImplement[T, Y](key, m[key]))
// 	}
// 	return result
// }

// RandomLoadBalancer 是一个通过随机算法从提供的运输函数列表中选择一个来执行HTTP请求的负载均衡器。
// 参数：
// - roundTripper：一个包含多个http.RoundTripper函数的切片，这些函数将被用于发送HTTP请求。
// - req：指向待发送的HTTP请求的指针。
// 返回值：
// - *http.Response：从运输函数中返回的HTTP响应指针，如果所有运输函数都失败，则为nil。
// - error：如果在发送请求时遇到错误，则返回错误信息；否则为nil。
// func RandomLoadBalancer(roundTripper map[string]func(*http.Request) (*http.Response, error), req *http.Request, upStreamServerSchemeAndHostOfName map[string]generic.PairInterface[string, string]) (*http.Response, error) {
// 	// 打印传入的运输函数列表
// 	log.Println("接收到的可用上游服务器:", roundTripper)

// 	PrintRequest(req)
// 	var roundTripperArray = mapToArray(roundTripper)
// 	// 使用随机算法对运输函数列表进行洗牌，以实现随机选择运输函数的效果
// 	var healthRoundTripper = randomShuffle(roundTripperArray)
// 	var rer error = nil

// 	// 遍历洗牌后的运输函数列表，尝试发送HTTP请求
// 	for _, transport := range healthRoundTripper {
// 		var name = transport.GetFirst()
// 		var Scheme = upStreamServerSchemeAndHostOfName[name].GetFirst()
// 		var Host = upStreamServerSchemeAndHostOfName[name].GetSecond()
// 		req.URL.Scheme = Scheme
// 		req.Host = Host
// 		req.URL.Host = Host
// 		var rs, err = transport.GetSecond()(req) // 执行运输函数
// 		if err != nil {
// 			// 如果请求发送失败，打印错误信息，并更新错误变量
// 			log.Println("ERROR:", err)
// 			rer = err

// 		} else if rs.StatusCode >= 500 {
// 			// 如果请求发送成功，打印响应信息，并返回响应和错误
// 			log.Println("ERROR:", "Status code is 500 or greater.", rs.StatusCode)
// 			PrintResponse(rs)
// 			rer =  errors.New("ERROR:Status code is 500 or greater" + fmt.Sprint(rs.StatusCode))
// 		} else {
// 			// 如果请求发送成功，打印响应信息，并返回响应和错误
// 			PrintResponse(rs)
// 			return rs, err
// 		}

// 	}
// 	// 如果所有运输函数尝试都失败，返回nil和错误信息
// 	return nil, rer
// }

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
