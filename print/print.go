package print

import (
	"log"
	"net/http"
)

func PrintRequest(req *http.Request) {
	// 打印HTTP请求的基本信息
	log.Println(" HTTP Request {")
	log.Printf("Method: %s\n", req.Method) // 打印请求方法
	log.Printf("URL: %s\n", req.URL)       // 打印请求的URL
	log.Printf("Proto: %s\n", req.Proto)   // 打印请求的协议版本
	log.Printf("host: %s\n", req.Host)

	// 打印请求头信息
	log.Printf("Header: \n")
	PrintHeader(req.Header)

	log.Println("} HTTP Request ")
}

// PrintResponse 打印HTTP响应的详细信息
// 参数：
// resp *http.Response: 一个指向http.Response的指针，包含了HTTP响应的全部信息
func PrintResponse(resp *http.Response) {
	// 打印HTTP响应的起始标志
	log.Println(" HTTP Response {")
	// 打印响应的状态信息
	log.Printf("Status: %s\n", resp.Status)
	// 打印响应的状态码
	log.Printf("StatusCode: %d\n", resp.StatusCode)
	// 打印响应的协议版本
	log.Printf("Proto: %s\n", resp.Proto)
	// 打印响应的头部信息
	log.Printf("Header: \n")
	PrintHeader(resp.Header)

	// 打印HTTP响应的结束标志
	log.Println("} HTTP Response ")
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
	log.Println(" HTTP Header {")
	// 遍历头部信息，并打印每一条键值对
	for key, values := range header {
		log.Printf("%s: %v\n", key, values)
	}
	// 打印HTTP头部结束标签
	log.Println("} HTTP Header ")
}
