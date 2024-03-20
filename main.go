package main

import (
	"crypto/tls"
	"fmt"
	"log"

	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// checkUpstreamHealth 检查上游服务的健康状态
// url: 要检查的服务的URL地址
// RoundTrip: 用于发送HTTP请求的函数，模拟HTTP客户端的行为
// 返回值: 返回一个布尔值，表示上游服务的健康状态，true为健康，false为不健康
func checkUpstreamHealth(url string, RoundTrip func(*http.Request) (*http.Response, error)) bool {

	statusCode, err := sendHeadRequestAndCheckStatus(url, RoundTrip)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}

	fmt.Printf("health check Status code: %d\n", statusCode)

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
func main() {

	upstreamServer := "https://www.example.com/"

	upstreamURL, err := url.Parse(upstreamServer)
	if err != nil {
		log.Fatalf("Failed to parse upstream server URL: %v", err)
	}
	if upstreamURL.Path != "/" && upstreamURL.Path != "" {
		log.Fatalf("upstreamServer Path must be / or empty")
	}

	http3Client := &http.Client{
		Transport: &http3.RoundTripper{
			TLSClientConfig: &tls.Config{},
			QuicConfig:      &quic.Config{},
		},
	}
	http2Client := &http.Client{
		Transport: http.DefaultTransport,
	}
	var transportsUpstream = []func(*http.Request) (*http.Response, error){
		func(req *http.Request) (*http.Response, error) { return http3Client.Transport.RoundTrip(req) },

		func(req *http.Request) (*http.Response, error) { return http2Client.Transport.RoundTrip(req) },
	}
	var maxAge = 30 * 1000
	var expires = int64(0)
	var healthyUpstream = transportsUpstream
	customTransport := &customRoundTripperLoadBalancer{
		upstreamURL: upstreamURL,
		getTransportHealthy: func() []func(*http.Request) (*http.Response, error) {
			if expires > time.Now().Unix() {
				fmt.Println("不需要进行健康检查")
				fmt.Println("healthyUpstream", healthyUpstream)
				return healthyUpstream
			}
			go func() {
				var healthy = []func(*http.Request) (*http.Response, error){}
				fmt.Println("需要进行健康检查")
				//进行健康检查
				for _, roundTrip := range transportsUpstream {
					if checkUpstreamHealth(upstreamServer, roundTrip) {
						healthy = append(healthy, roundTrip)
					}
				}
				if len(healthy) == 0 {
					healthyUpstream = transportsUpstream
				} else {
					healthyUpstream = healthy
					fmt.Println("healthyUpstream", healthyUpstream)
				}
				expires = time.Now().Unix() + int64(maxAge)
			}()
			return healthyUpstream
		},
	}

	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)

	proxy.Transport = customTransport

	server := &http.Server{
		Addr:    ":18080",
		Handler: proxy,
	}

	log.Printf("Starting reverse proxy server on :18080")
	log.Fatal(server.ListenAndServeTLS("cert.crt", "key.pem"))
}

type customRoundTripperLoadBalancer struct {
	upstreamURL         *url.URL
	getTransportHealthy func() []func(*http.Request) (*http.Response, error)
}

func (c *customRoundTripperLoadBalancer) RoundTrip(req *http.Request) (*http.Response, error) {

	req.Host = c.upstreamURL.Host

	PrintRequest(req)
	var rs, err = RandomLoadBalancer(c.getTransportHealthy(), req)

	if err != nil {
		fmt.Println("ERROR:", err)
	} else {
		PrintResponse(rs)
	}
	return rs, err
}

func randomShuffle[T any](arr []T) []T {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
	return arr
}
func RandomLoadBalancer(roundTripper []func(*http.Request) (*http.Response, error), req *http.Request) (*http.Response, error) {
	fmt.Println("RoundTripper:", roundTripper)

	var healthRoundTripper = randomShuffle(roundTripper)
	var rer error = nil

	for _, transport := range healthRoundTripper {
		var rs, err = transport(req)
		if err != nil {
			fmt.Println("ERROR:", err)
			rer = err

		} else {
			PrintResponse(rs)
			return rs, err
		}

	}
	return nil, rer
}

func PrintRequest(req *http.Request) {
	fmt.Println(" HTTP Request {")
	fmt.Printf("Method: %s\n", req.Method)
	fmt.Printf("URL: %s\n", req.URL)
	fmt.Printf("Proto: %s\n", req.Proto)

	fmt.Printf("Header: \n")
	PrintHeader(req.Header)

	fmt.Println("} HTTP Request ")
}

func PrintResponse(resp *http.Response) {
	fmt.Println(" HTTP Response {")
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("StatusCode: %d\n", resp.StatusCode)
	fmt.Printf("Proto: %s\n", resp.Proto)
	fmt.Printf("Header: \n")
	PrintHeader(resp.Header)

	fmt.Println("} HTTP Response ")
}

func PrintHeader(header http.Header) {
	fmt.Println(" HTTP Header {")
	for key, values := range header {
		fmt.Printf("%s: %v\n", key, values)
	}
	fmt.Println("} HTTP Header ")
}
