package h3

import (
	// "context"
	// "crypto/tls"

	"io"
	"log"

	// "net"
	// "net/http"
	"testing"
)

// TestHttpViaIP 通过指定的IP地址对给定的URL进行HTTP请求测试。
//
// 参数:
//
//	t *testing.T - 测试框架提供的测试上下文，用于报告测试失败和日志记录。
//
// 无返回值。

// TestHttpsViaIP 通过指定的IP地址测试HTTPS连接。
// 参数:
//
//	t *testing.T - 测试框架提供的测试上下文，用于报告测试失败和日志记录。
//
// 无返回值。
func TestHttp3ViaIP(t *testing.T) {
	var addresses = []string{"104.21.43.14"}
	// var eee error = nil
	var failure = 0
	var success = 0
	for _, address := range addresses {
		ip := address                                                                 // 要访问的目标IP地址
		url := "https://production.hello-word-worker-cloudflare.masx200.workers.dev/" // 要访问的目标URL

		// 使用指定的IP地址发起HTTP GET请求
		resp, err := FetchHttp3WithIP(ip, url)
		if err != nil {
			log.Println("Error:", err)
			// t.Errorf(err.Error())
			// eee = err
			failure += 1
			continue
		}
		PrintResponse(resp)
		// 确保响应体在函数返回前被关闭
		defer resp.Body.Close()
		// 读取并打印响应体内容
		body, _ := io.ReadAll(resp.Body)
		log.Println("Response:", string(body))
		success += 1
		continue
	}

	if failure > 0 {
		log.Println("Error:", "have failure test")
		// t.Errorf("No successful test")
	}
	if success == 0 {
		log.Println("Error:", "No successful test")
		t.Errorf("No successful test")
	}
	log.Println("http3 Success:", success)
}
