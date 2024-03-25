package print

import (
	"fmt"
	"net/http"
)

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
